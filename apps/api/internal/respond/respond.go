package respond

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func Error(w http.ResponseWriter, status int, message, code string) {
	JSON(w, status, errorBody{Error: message, Code: code})
}

func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "an unexpected error occurred", "internal_error")
}
