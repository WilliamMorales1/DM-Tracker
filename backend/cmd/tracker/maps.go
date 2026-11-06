package main

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "golang.org/x/image/webp"
)

const maxMapBytes = 16 << 20

var mapNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type gameMap struct {
	ID  string `json:"id"`
	Img string `json:"img"`
}

func listGameMaps(ctx context.Context) ([]gameMap, error) {
	rows, err := conn.QueryContext(ctx, `SELECT id, img FROM GameMaps ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	maps := make([]gameMap, 0)
	for rows.Next() {
		var gm gameMap
		if err := rows.Scan(&gm.ID, &gm.Img); err != nil {
			return nil, err
		}
		maps = append(maps, gm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return maps, nil
}

func handleAPIMaps(w http.ResponseWriter, r *http.Request) {
	maps, err := listGameMaps(r.Context())
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not list maps", err)
		return
	}
	writeJSON(w, maps)
}

func handleAPICreateMap(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMapBytes+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		httpError(w, r, http.StatusBadRequest, "invalid map upload", err)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if !mapNamePattern.MatchString(name) {
		httpError(w, r, http.StatusBadRequest, "map name must use letters, numbers, _ or -", nil)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		httpError(w, r, http.StatusBadRequest, "map image required", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxMapBytes+1))
	if err != nil {
		httpError(w, r, http.StatusBadRequest, "could not read map image", err)
		return
	}
	if len(data) > maxMapBytes {
		httpError(w, r, http.StatusRequestEntityTooLarge, "map image exceeds 16 MiB", nil)
		return
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 {
		httpError(w, r, http.StatusBadRequest, "unsupported or invalid map image", err)
		return
	}

	extension := map[string]string{"jpeg": ".jpg", "png": ".png", "webp": ".webp"}[format]
	if extension == "" {
		httpError(w, r, http.StatusBadRequest, "map image must be JPEG, PNG, or WebP", nil)
		return
	}
	if err := os.MkdirAll("maps", 0o755); err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not save map", err)
		return
	}
	filename := name + extension
	path := filepath.Join("maps", filename)
	imgURL := "/maps/" + filename

	tx, err := conn.BeginTx(r.Context(), nil)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not save map", err)
		return
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO GameMaps (id, img) VALUES (?, ?)`, name, imgURL); err != nil {
		httpError(w, r, http.StatusConflict, "map name already exists", err)
		return
	}

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		httpError(w, r, http.StatusConflict, "map file already exists", err)
		return
	}
	if _, err := out.Write(data); err != nil {
		_ = out.Close()
		_ = os.Remove(path)
		httpError(w, r, http.StatusInternalServerError, "could not save map image", err)
		return
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(path)
		httpError(w, r, http.StatusInternalServerError, "could not save map image", err)
		return
	}
	if err := tx.Commit(); err != nil {
		_ = os.Remove(path)
		httpError(w, r, http.StatusInternalServerError, "could not save map", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, gameMap{ID: name, Img: imgURL})
}
