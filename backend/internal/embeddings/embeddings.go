// Package embeddings turns text into vector embeddings for semantic search
// via any OpenAI-compatible server's /embeddings endpoint.
package embeddings

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"tracker/internal/llm"
)

// DefaultModel is served under this id by both Ollama and Lemonade (user.nomic-embed-text-v2-moe).
const DefaultModel = "nomic-embed-text-v2-moe"

var client = &http.Client{Timeout: 60 * time.Second}

// Model returns LLM_EMBED_MODEL, or DefaultModel.
func Model() string {
	if m := os.Getenv("LLM_EMBED_MODEL"); m != "" {
		return m
	}
	return DefaultModel
}

// Embed returns the embedding for text as a JSON-array string (e.g.
// "[0.1,0.2,...]"), suitable for storing in a sqlite column.
func Embed(ctx context.Context, text string) (string, error) {
	var er struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := llm.PostJSON(ctx, client, "/embeddings", map[string]any{"model": Model(), "input": text}, &er); err != nil {
		return "", fmt.Errorf("embedding request: %w", err)
	}
	if len(er.Data) == 0 || len(er.Data[0].Embedding) == 0 {
		return "", fmt.Errorf("embedding request: empty response from model %q", Model())
	}
	parts := make([]string, len(er.Data[0].Embedding))
	for i, f := range er.Data[0].Embedding {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]", nil
}
