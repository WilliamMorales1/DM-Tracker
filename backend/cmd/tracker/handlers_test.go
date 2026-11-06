package main

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tracker/internal/db"
)

// setupTestDB points the package-level conn/roConn/q at a fresh migrated
// sqlite database in a temp dir, so handler tests exercise real queries
// without touching the dev database.
func setupTestDB(t *testing.T) {
	t.Helper()
	dbFile := filepath.Join(t.TempDir(), "test.db")

	if err := db.RunMigrations("../../migrations", dbFile); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	c, err := db.Open(dbFile)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { c.Close() })

	ro, err := db.OpenReadOnly(dbFile)
	if err != nil {
		t.Fatalf("open read-only db: %v", err)
	}
	t.Cleanup(func() { ro.Close() })

	conn = c
	roConn = ro
	q = db.New(c)
}

// chdirIntoTempNotesDir switches the process cwd to a temp dir containing a
// "notes" subdir, so handlers that write to ./notes (e.g. handleAPINote's
// PUT) don't touch the real backend/notes/ directory during tests. Must be
// called after setupTestDB, since that resolves migration paths relative to
// the original cwd.
func chdirIntoTempNotesDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, notesDir), 0755); err != nil {
		t.Fatalf("mkdir notes: %v", err)
	}
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestHandleAPICreateMapStoresUploadedImage(t *testing.T) {
	setupTestDB(t)
	workDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(workDir, "maps"), 0755); err != nil {
		t.Fatalf("mkdir maps: %v", err)
	}
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if err := form.WriteField("name", "Cavern"); err != nil {
		t.Fatal(err)
	}
	file, err := form.CreateFormFile("image", "cavern.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(file, &pngData); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/maps", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	w := httptest.NewRecorder()
	handleAPICreateMap(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join("maps", "Cavern.png")); err != nil {
		t.Fatalf("uploaded map missing: %v", err)
	}

	list := httptest.NewRecorder()
	handleAPIMaps(list, httptest.NewRequest(http.MethodGet, "/api/maps", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"id":"Cavern"`) {
		t.Fatalf("map missing from list: status=%d body=%s", list.Code, list.Body.String())
	}
}

func TestHandleAPIMonsterMissingID(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/monsters/999999", nil)
	req.SetPathValue("id", "999999")
	w := httptest.NewRecorder()
	handleAPIMonster(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleAPIMonsterInvalidID(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/monsters/not-a-number", nil)
	req.SetPathValue("id", "not-a-number")
	w := httptest.NewRecorder()
	handleAPIMonster(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleAPINoteInvalidName(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/x", nil)
	req.SetPathValue("name", "../../etc/passwd")
	w := httptest.NewRecorder()
	handleAPIGetNote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleAPINotePutGetRoundtrip(t *testing.T) {
	setupTestDB(t)
	chdirIntoTempNotesDir(t)

	put := httptest.NewRequest(http.MethodPut, "/api/notes/sesh-test.md", strings.NewReader("hello world"))
	put.SetPathValue("name", "sesh-test.md")
	w := httptest.NewRecorder()
	handleAPIPutNote(w, put)
	if w.Code != http.StatusNoContent {
		t.Fatalf("PUT status = %d, want %d; body: %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/api/notes/sesh-test.md", nil)
	get.SetPathValue("name", "sesh-test.md")
	w = httptest.NewRecorder()
	handleAPIGetNote(w, get)
	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "hello world") {
		t.Errorf("GET body missing saved content, got: %s", w.Body.String())
	}
}

func TestHandleAPIInitiativePresetRoundtrip(t *testing.T) {
	setupTestDB(t)

	body := `[{"name":"Goblin","init":12,"hp":7,"ac":15}]`
	put := httptest.NewRequest(http.MethodPut, "/api/initiative-presets/Ambush", strings.NewReader(body))
	put.SetPathValue("name", "Ambush")
	w := httptest.NewRecorder()
	handleAPIPutInitiativePreset(w, put)
	if w.Code != http.StatusNoContent {
		t.Fatalf("PUT status = %d, want %d; body: %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	list := httptest.NewRequest(http.MethodGet, "/api/initiative-presets", nil)
	w = httptest.NewRecorder()
	handleAPIInitiativePresets(w, list)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Goblin") {
		t.Errorf("list missing saved preset, got: %s", w.Body.String())
	}

	del := httptest.NewRequest(http.MethodDelete, "/api/initiative-presets/Ambush", nil)
	del.SetPathValue("name", "Ambush")
	w = httptest.NewRecorder()
	handleAPIDeleteInitiativePreset(w, del)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleAPIInitiativePresetInvalidJSON(t *testing.T) {
	setupTestDB(t)

	put := httptest.NewRequest(http.MethodPut, "/api/initiative-presets/Bad", strings.NewReader("not json"))
	put.SetPathValue("name", "Bad")
	w := httptest.NewRecorder()
	handleAPIPutInitiativePreset(w, put)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleAPIChatMissingParams(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	w := httptest.NewRecorder()
	handleAPIChat(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing q,model: status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/chat?q=hello", nil)
	w = httptest.NewRecorder()
	handleAPIChat(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing model: status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// execSQL is the chat agent's SQL tool; it must reject writes even though the
// connection it's handed (roConn) is opened with query_only, and even when a
// SELECT-prefixed statement smuggles a write in a later clause.
func TestExecSQLRejectsNonSelect(t *testing.T) {
	setupTestDB(t)

	out := execSQL(t.Context(), "DELETE FROM Sessions")
	if !strings.Contains(out, "only SELECT") {
		t.Errorf("expected rejection message, got: %s", out)
	}
}

func TestExecSQLReadOnlyConnectionBlocksWrites(t *testing.T) {
	setupTestDB(t)

	out := execSQL(t.Context(), "SELECT 1; DELETE FROM Sessions")
	if strings.Contains(out, "0 |") || !strings.Contains(strings.ToLower(out), "error") {
		t.Errorf("expected query_only connection to error on smuggled write, got: %s", out)
	}
}

func TestHandleAPIOllamaModelsNoOllamaRunning(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/ollama-models", nil)
	w := httptest.NewRecorder()
	handleAPIOllamaModels(w, req)

	// In CI/sandboxed environments there's no local Ollama; the handler
	// should degrade to a clean 502 rather than panicking or hanging.
	if w.Code != http.StatusBadGateway && w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d or %d", w.Code, http.StatusBadGateway, http.StatusOK)
	}
}
