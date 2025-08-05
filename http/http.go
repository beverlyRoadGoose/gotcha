package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	Authorization = "Authorization"

	ContentType                          = "Content-Type"
	ContentTypeApplicationJson           = "application/json"
	ContentTypeApplicationFormUrlEncoded = "application/x-www-form-urlencoded"
)

// ErrorResponse represents a standardized error response format
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func ParseRequestBody[T any](r *http.Request) (*T, int, error) {
	if r.Body == nil {
		return nil, http.StatusBadRequest, errors.New("request body is nil")
	}

	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(r.Body)

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var request T
	err = json.Unmarshal(requestBody, &request)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "failed to unmarshal request body: %s", string(requestBody[:]))
	}

	return &request, 0, nil
}

func JSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Add(ContentType, ContentTypeApplicationJson)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func RespondWithError(statusCode int, w http.ResponseWriter, response interface{}, contentType string) {
	w.Header().Add(ContentType, contentType)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func HandlePanic(ctx context.Context, w http.ResponseWriter) []byte {
	if r := recover(); r != nil {
		stackTrace := make([]byte, 64*1024) // Buffer to hold the stack trace

		RespondWithError(http.StatusInternalServerError, w, ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal server error",
		}, ContentTypeApplicationJson)

		return stackTrace
	}

	return nil
}
