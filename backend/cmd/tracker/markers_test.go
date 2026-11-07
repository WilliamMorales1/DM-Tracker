package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func markerReq(method, body string, id, i string) *http.Request {
	req := httptest.NewRequest(method, "/api/maps/"+id+"/markers", strings.NewReader(body))
	req.SetPathValue("id", id)
	req.SetPathValue("i", i)
	return req
}

func TestMapMarkerCRUD(t *testing.T) {
	setupTestDB(t)
	if _, err := conn.Exec(`INSERT INTO GameMaps (id, img) VALUES ('Cave', '/maps/Cave.png')`); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handleAPICreateMapMarker(w, markerReq("POST", `{"x":10,"y":20,"title":"Door","body":"locked"}`, "Cave", ""))
	if w.Code != http.StatusCreated || !strings.Contains(w.Body.String(), `"i":1`) {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	handleAPICreateMapMarker(w, markerReq("POST", `{"x":1,"y":2,"title":"Second"}`, "Cave", ""))
	if !strings.Contains(w.Body.String(), `"i":2`) {
		t.Fatalf("second id: %s", w.Body)
	}

	w = httptest.NewRecorder()
	handleAPICreateMapMarker(w, markerReq("POST", `{"x":1,"y":2,"title":" "}`, "Cave", ""))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("blank title: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleAPICreateMapMarker(w, markerReq("POST", `{"x":1,"y":2,"title":"x"}`, "Nope", ""))
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown map: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleAPIUpdateMapMarker(w, markerReq("PUT", `{"x":5,"y":6,"title":"Gate","body":"open"}`, "Cave", "1"))
	if w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body)
	}

	w = httptest.NewRecorder()
	handleAPIMapMarkers(w, markerReq("GET", "", "Cave", ""))
	if !strings.Contains(w.Body.String(), "Gate") || !strings.Contains(w.Body.String(), "Second") {
		t.Fatalf("list: %s", w.Body)
	}

	w = httptest.NewRecorder()
	handleAPIDeleteMapMarker(w, markerReq("DELETE", "", "Cave", "1"))
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleAPIDeleteMapMarker(w, markerReq("DELETE", "", "Cave", "1"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("redelete: %d", w.Code)
	}
}
