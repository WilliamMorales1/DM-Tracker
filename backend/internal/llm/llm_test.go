package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// Response bodies captured from real Lemonade 13305/api/v1 and Ollama 11434/v1 servers (trimmed).
const lemonadeModels = `{"data":[
 {"id":"Qwen3.5-4B","labels":["chat","custom","reasoning","tool-calling"],"object":"model"},
 {"id":"nomic-embed-text-v2-moe","labels":["custom","embeddings"],"object":"model"},
 {"id":"Whisper-Tiny","labels":["transcription"],"object":"model"},
 {"id":"ollama.gemma4:31b","labels":["cloud","chat"],"object":"model"}]}`

const ollamaModels = `{"object":"list","data":[
 {"id":"gemma4:31b-cloud","object":"model","created":1,"owned_by":"library"},
 {"id":"nomic-embed-text-v2-moe:latest","object":"model","created":1,"owned_by":"library"},
 {"id":"qwen3.5:4b","object":"model","created":1,"owned_by":"library"}]}`

func fakeServer(t *testing.T, prefix, models string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+prefix+"/models", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(models))
	})
	mux.HandleFunc("POST "+prefix+"/embeddings", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Model string `json:"model"`
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Input == "" {
			http.Error(w, "bad input", http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"object":"list","model":"` + in.Model + `","data":[{"object":"embedding","index":0,"embedding":[0.5,-0.25]}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func useCandidates(t *testing.T, c ...string) {
	t.Helper()
	old := Candidates
	Candidates = c
	forget()
	t.Cleanup(func() { Candidates = old; forget() })
}

func TestChatModelsFiltersPerServer(t *testing.T) {
	cases := []struct {
		name, prefix, models string
		want                 []string
	}{
		{"lemonade", "/api/v1", lemonadeModels, []string{"Qwen3.5-4B", "ollama.gemma4:31b"}},
		{"ollama", "/v1", ollamaModels, []string{"gemma4:31b-cloud", "qwen3.5:4b"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := fakeServer(t, tc.prefix, tc.models)
			t.Setenv("LLM_BASE_URL", srv.URL+tc.prefix+"/")
			got, err := ChatModels(t.Context(), srv.Client())
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBaseURLProbesInOrder(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	ollama := fakeServer(t, "/v1", ollamaModels)
	// Dead Lemonade candidate first: probe must fall through to Ollama.
	useCandidates(t, "http://127.0.0.1:1/api/v1", ollama.URL+"/v1")
	got, err := BaseURL(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got != ollama.URL+"/v1" {
		t.Fatalf("got %q, want ollama", got)
	}
}

func TestBaseURLNoneReachable(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	useCandidates(t, "http://127.0.0.1:1/v1")
	if _, err := BaseURL(t.Context()); err == nil {
		t.Fatal("want error when no server reachable")
	}
}

func TestPostJSONEmbeddingsBothServers(t *testing.T) {
	for _, prefix := range []string{"/api/v1", "/v1"} {
		srv := fakeServer(t, prefix, "{}")
		t.Setenv("LLM_BASE_URL", srv.URL+prefix)
		var out struct {
			Data []struct {
				Embedding []float32 `json:"embedding"`
			} `json:"data"`
		}
		if err := PostJSON(t.Context(), srv.Client(), "/embeddings", map[string]any{"model": "m", "input": "x"}, &out); err != nil {
			t.Fatalf("%s: %v", prefix, err)
		}
		if len(out.Data) != 1 || !slices.Equal(out.Data[0].Embedding, []float32{0.5, -0.25}) {
			t.Fatalf("%s: got %+v", prefix, out)
		}
	}
}

func TestPostJSONErrorStatus(t *testing.T) {
	srv := fakeServer(t, "/v1", "{}")
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")
	var out any
	err := PostJSON(t.Context(), srv.Client(), "/embeddings", map[string]any{"model": "m"}, &out)
	if err == nil {
		t.Fatal("want error on 400")
	}
}
