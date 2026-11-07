package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const (
	maxMarkerBody  = 500
	maxMarkerTitle = 100
)

type mapMarker struct {
	I     int    `json:"i"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func decodeMarker(w http.ResponseWriter, r *http.Request) (mapMarker, bool) {
	var m mapMarker
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		httpError(w, r, http.StatusBadRequest, "invalid marker", err)
		return m, false
	}
	m.Title = strings.TrimSpace(m.Title)
	switch {
	case m.Title == "" || len([]rune(m.Title)) > maxMarkerTitle:
		httpError(w, r, http.StatusBadRequest, "marker title required (max 100 characters)", nil)
	case len([]rune(m.Body)) > maxMarkerBody:
		httpError(w, r, http.StatusBadRequest, "marker description exceeds 500 characters", nil)
	case m.X < 0 || m.Y < 0:
		httpError(w, r, http.StatusBadRequest, "marker position must be non-negative", nil)
	default:
		return m, true
	}
	return m, false
}

func handleAPIMapMarkers(w http.ResponseWriter, r *http.Request) {
	rows, err := conn.QueryContext(r.Context(),
		`SELECT i, x, y, title, body FROM MapMarkers WHERE map_id = ? ORDER BY i`, r.PathValue("id"))
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not list markers", err)
		return
	}
	defer rows.Close()
	markers := make([]mapMarker, 0)
	for rows.Next() {
		var m mapMarker
		if err := rows.Scan(&m.I, &m.X, &m.Y, &m.Title, &m.Body); err != nil {
			httpError(w, r, http.StatusInternalServerError, "could not list markers", err)
			return
		}
		markers = append(markers, m)
	}
	if err := rows.Err(); err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not list markers", err)
		return
	}
	writeJSON(w, markers)
}

func handleAPICreateMapMarker(w http.ResponseWriter, r *http.Request) {
	m, ok := decodeMarker(w, r)
	if !ok {
		return
	}
	err := conn.QueryRowContext(r.Context(),
		`INSERT INTO MapMarkers (map_id, i, x, y, title, body)
		 SELECT ?1, COALESCE(MAX(i), 0) + 1, ?2, ?3, ?4, ?5 FROM MapMarkers WHERE map_id = ?1
		 RETURNING i`,
		r.PathValue("id"), m.X, m.Y, m.Title, m.Body).Scan(&m.I)
	if err != nil {
		httpError(w, r, http.StatusNotFound, "could not add marker", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, m)
}

func handleAPIUpdateMapMarker(w http.ResponseWriter, r *http.Request) {
	i, err := strconv.Atoi(r.PathValue("i"))
	if err != nil {
		httpError(w, r, http.StatusBadRequest, "invalid marker id", err)
		return
	}
	m, ok := decodeMarker(w, r)
	if !ok {
		return
	}
	m.I = i
	res, err := conn.ExecContext(r.Context(),
		`UPDATE MapMarkers SET x = ?, y = ?, title = ?, body = ? WHERE map_id = ? AND i = ?`,
		m.X, m.Y, m.Title, m.Body, r.PathValue("id"), i)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not update marker", err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, r, http.StatusNotFound, "marker not found", sql.ErrNoRows)
		return
	}
	writeJSON(w, m)
}

func handleAPIDeleteMapMarker(w http.ResponseWriter, r *http.Request) {
	i, err := strconv.Atoi(r.PathValue("i"))
	if err != nil {
		httpError(w, r, http.StatusBadRequest, "invalid marker id", err)
		return
	}
	res, err := conn.ExecContext(r.Context(),
		`DELETE FROM MapMarkers WHERE map_id = ? AND i = ?`, r.PathValue("id"), i)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "could not delete marker", err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, r, http.StatusNotFound, "marker not found", errors.New("no rows"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
