// Package llm talks to any OpenAI-compatible server (Lemonade, Ollama, llama.cpp, vLLM, ...).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// Candidates are probed in order when LLM_BASE_URL is unset.
var Candidates = []string{
	"http://localhost:13305/api/v1", // Lemonade
	"http://localhost:11434/v1",     // Ollama
}

var probeClient = &http.Client{Timeout: 3 * time.Second}

var (
	mu       sync.Mutex
	resolved string
)

// BaseURL returns LLM_BASE_URL, or the first reachable entry in Candidates.
func BaseURL(ctx context.Context) (string, error) {
	if v := strings.TrimRight(os.Getenv("LLM_BASE_URL"), "/"); v != "" {
		return v, nil
	}
	mu.Lock()
	defer mu.Unlock()
	if resolved != "" {
		return resolved, nil
	}
	var errs []error
	for _, base := range Candidates {
		if _, err := listModels(ctx, probeClient, base); err != nil {
			errs = append(errs, err)
			continue
		}
		resolved = base
		return base, nil
	}
	return "", fmt.Errorf("no LLM server reachable (set LLM_BASE_URL or start Lemonade/Ollama): %w", errors.Join(errs...))
}

// forget drops the cached server so the next call re-probes, e.g. after a server swap.
func forget() {
	mu.Lock()
	resolved = ""
	mu.Unlock()
}

func setAuth(req *http.Request) {
	if key := os.Getenv("LLM_API_KEY"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
}

// Model is one entry from GET /models.
type Model struct {
	ID string `json:"id"`
	// Labels is Lemonade-specific ("chat", "embeddings", ...); empty on other servers.
	Labels []string `json:"labels"`
}

func listModels(ctx context.Context, client *http.Client, base string) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/models", nil)
	if err != nil {
		return nil, err
	}
	setAuth(req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s/models: %s", base, resp.Status)
	}
	var mr struct {
		Data []Model `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, fmt.Errorf("decode %s/models: %w", base, err)
	}
	return mr.Data, nil
}

// ChatModels returns ids of models usable for chat, skipping embedding/rerank/speech models.
func ChatModels(ctx context.Context, client *http.Client) ([]string, error) {
	base, err := BaseURL(ctx)
	if err != nil {
		return nil, err
	}
	models, err := listModels(ctx, client, base)
	if err != nil {
		forget()
		return nil, fmt.Errorf("LLM server not reachable - is it running? (%w)", err)
	}
	ids := []string{}
	for _, m := range models {
		if len(m.Labels) > 0 {
			if slices.Contains(m.Labels, "chat") {
				ids = append(ids, m.ID)
			}
			continue
		}
		// Ollama lists embedding models alongside chat models with no labels.
		if !strings.Contains(strings.ToLower(m.ID), "embed") {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

// PostJSON sends in to base+path and decodes the response into out.
func PostJSON(ctx context.Context, client *http.Client, path string, in, out any) error {
	base, err := BaseURL(ctx)
	if err != nil {
		return err
	}
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal %s request: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	setAuth(req)
	resp, err := client.Do(req)
	if err != nil {
		forget()
		return fmt.Errorf("LLM server at %s not reachable or timed out - is it running? (%w)", base, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s%s: %s %s", base, path, resp.Status, strings.TrimSpace(buf.String()))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}
	return nil
}
