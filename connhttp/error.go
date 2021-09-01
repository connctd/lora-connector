package connhttp

import (
	"encoding/json"
	"net/http"
)

var (
	ErrorBadContentType    = Error{Code: http.StatusBadRequest, Message: "Expected content type to be application/json"}
	ErrorMissingInstanceID = Error{Code: http.StatusBadRequest, Message: "Header X-Instance-ID is missing"}
	ErrorBadRequestBody    = Error{Code: http.StatusBadRequest, Message: "Empty or malformed request body"}
	ErrorInvalidJsonBody   = Error{Code: http.StatusBadRequest, Message: "Request body does not contain valid json"}
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) WriteBody(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	b, err := json.Marshal(e)
	if err != nil {
		w.Write([]byte("{\"err\":\"Failed to marshal error\"}"))
	} else {
		w.Write(b)
	}
}

func (e Error) Error() string {
	return e.Message
}
