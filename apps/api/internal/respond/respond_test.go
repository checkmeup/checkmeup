package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_WritesStatusAndBody(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusCreated, map[string]string{"hello": "world"})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got["hello"] != "world" {
		t.Fatalf("body = %v, want hello=world", got)
	}
}

func TestJSON_MarshalFailureFallsBackToPlainTextError(t *testing.T) {
	w := httptest.NewRecorder()

	// channels can't be marshaled to JSON, forcing json.Marshal to fail.
	JSON(w, http.StatusOK, make(chan int))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if ct := w.Header().Get("Content-Type"); ct == "application/json" {
		t.Fatalf("Content-Type should not be application/json on marshal failure, got %q", ct)
	}
}

func TestError_WritesErrorBody(t *testing.T) {
	w := httptest.NewRecorder()

	Error(w, http.StatusBadRequest, "missing field", "validation_error")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var got errorBody
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Error != "missing field" || got.Code != "validation_error" {
		t.Fatalf("body = %+v, want {missing field validation_error}", got)
	}
}

func TestInternalError_WritesGenericMessage(t *testing.T) {
	w := httptest.NewRecorder()

	InternalError(w)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var got errorBody
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Error != "an unexpected error occurred" || got.Code != "internal_error" {
		t.Fatalf("body = %+v, want {an unexpected error occurred internal_error}", got)
	}
}
