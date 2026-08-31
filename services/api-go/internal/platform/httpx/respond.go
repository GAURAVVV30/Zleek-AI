// Package httpx provides the standard JSON envelope used by every HTTP handler
// so the React UI (which unwraps response.data and reads data.<field>, and the
// auth context which reads data.user/accessToken) gets one consistent shape:
//
//	Success: { "success": true,  "data": { ... } }
//	Error:   { "success": false, "error": { "code": <http status>, "message": "..." } }
package httpx

import (
	"encoding/json"
	"net/http"
)

func Envelope(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    status,
			"message": message,
		},
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
