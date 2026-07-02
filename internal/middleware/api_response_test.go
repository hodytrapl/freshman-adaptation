package middleware_test

import (
	"encoding/json"
	"errors"
	"freshman-adaptation/internal/apperrors"
	"freshman-adaptation/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	middleware.WriteJSON(recorder, http.StatusCreated, map[string]string{"status": "ok"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("content-type = %q", contentType)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()

	middleware.WriteError(recorder, apperrors.NotFound("task not found"))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload.Error != "task not found" {
		t.Fatalf("error = %q", payload.Error)
	}
}

func TestWriteErrorUnknown(t *testing.T) {
	recorder := httptest.NewRecorder()

	middleware.WriteError(recorder, errors.New("unexpected"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestRecover(t *testing.T) {
	handler := middleware.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}
