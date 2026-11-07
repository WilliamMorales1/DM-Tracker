package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tracker/internal/db"
)

var conn *sql.DB
var roConn *sql.DB
var q *db.Queries

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

const dbPath = "tracker.db"

const frontendDist = "../frontend"

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(frontendDist, filepath.Clean(r.URL.Path))
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	http.ServeFile(w, r, filepath.Join(frontendDist, "index.html"))
}

func main() {
	if err := db.RunMigrations("migrations", dbPath); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	var err error
	conn, err = db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	roConn, err = db.OpenReadOnly(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer roConn.Close()
	q = db.New(conn)
	if err := syncNotesFromDisk(ctx); err != nil {
		log.Fatalf("sync notes: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	mux.Handle("GET /maps/", http.StripPrefix("/maps/", http.FileServer(http.Dir("maps"))))
	mux.HandleFunc("GET /api/monsters", handleAPIMonsters)
	mux.HandleFunc("GET /api/monsters/{id}", handleAPIMonster)
	mux.HandleFunc("GET /api/monster-stats", handleAPIMonsterStats)
	mux.HandleFunc("GET /api/spells", handleAPISpells)
	mux.HandleFunc("GET /api/spells/{id}", handleAPISpell)
	mux.HandleFunc("GET /api/actions", handleAPIActions)
	mux.HandleFunc("GET /api/skill-areas", handleAPISkillAreas)
	mux.HandleFunc("GET /api/conditions", handleAPIConditions)
	mux.HandleFunc("GET /api/exhaustion-levels", handleAPIExhaustionLevels)
	mux.HandleFunc("GET /api/maps", handleAPIMaps)
	mux.HandleFunc("POST /api/maps", handleAPICreateMap)
	mux.HandleFunc("GET /api/maps/{id}/markers", handleAPIMapMarkers)
	mux.HandleFunc("POST /api/maps/{id}/markers", handleAPICreateMapMarker)
	mux.HandleFunc("PUT /api/maps/{id}/markers/{i}", handleAPIUpdateMapMarker)
	mux.HandleFunc("DELETE /api/maps/{id}/markers/{i}", handleAPIDeleteMapMarker)
	mux.HandleFunc("GET /api/chat", handleAPIChat)
	mux.HandleFunc("GET /api/ollama-models", handleAPIOllamaModels)
	mux.HandleFunc("GET /api/search", handleAPISearch)
	mux.HandleFunc("GET /api/notes", handleAPINotesList)
	mux.HandleFunc("GET /api/notes/{name}", handleAPIGetNote)
	mux.HandleFunc("PUT /api/notes/{name}", handleAPIPutNote)
	mux.HandleFunc("GET /api/initiative-presets", handleAPIInitiativePresets)
	mux.HandleFunc("PUT /api/initiative-presets/{name}", handleAPIPutInitiativePreset)
	mux.HandleFunc("DELETE /api/initiative-presets/{name}", handleAPIDeleteInitiativePreset)
	mux.HandleFunc("/", serveFrontend)

	server := &http.Server{
		Addr:              "localhost:8080",
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	log.Println("Listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
