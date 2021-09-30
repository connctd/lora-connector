package api

import (
	"encoding/json"
	"net/http"
)

// NewError constructs an error
func NewError(err string, description string, status int) *Error {
	return &Error{
		APIError:    err,
		Status:      status,
		Description: description,
	}
}

// Error defines an error
type Error struct {
	APIError    string `json:"error"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// Error returns the errors description
func (e *Error) Error() string {
	return e.Description
}

// Write uses given response writer to write an error
func (e *Error) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Status)

	b, err := json.Marshal(e)
	if err != nil {
		w.Write([]byte("{\"error\":\"" + err.Error() + "\"}"))
	}

	w.Write(b)
}
