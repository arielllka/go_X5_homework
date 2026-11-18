package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

const MaxBodyBytes = 1024

type inputError struct {
	status        int
	message, code string
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) *inputError {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &inputError{415, "Content-Type must be application/json", "UNSUPPORTED_MEDIA_TYPE"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return decodeError(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err != nil {
			return decodeError(err)
		}
		return &inputError{400, "body must contain one JSON object", "INVALID_JSON"}
	}
	return nil
}
func decodeError(err error) *inputError {
	var large *http.MaxBytesError
	if errors.As(err, &large) {
		return &inputError{413, "body exceeds 1024 bytes", "BODY_TOO_LARGE"}
	}
	return &inputError{400, "body must contain one JSON object with known fields", "INVALID_JSON"}
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		status = 500
		data = []byte(`{"error":"internal server error","code":"INTERNAL_ERROR"}`)
		w.Header().Del("Location")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(data, '\n'))
}
func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, ErrorResponse{Error: message, Code: code})
}
