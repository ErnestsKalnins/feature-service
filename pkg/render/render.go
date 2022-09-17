// Package render provides convenience methods for rendering JSON payloads via
// http.ResponseWriter.
package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

// Error renders an HTTP response with the given status, and formats error as
// JSON.
func Error(w http.ResponseWriter, err error) {
	w.WriteHeader(codeFromErr(err))
	JSON(w, errorResponse{Error: err.Error()})
}

func codeFromErr(err error) int {
	for {
		if err == nil {
			return http.StatusInternalServerError
		}
		if c, ok := err.(interface{ Code() int }); ok {
			return c.Code()
		}
		err = errors.Unwrap(err)
	}
}

// JSON renders the given value as JSON.
func JSON(w http.ResponseWriter, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(buf.Bytes())
}
