package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse is the response returned to the user.
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// Error represents the error returned to the user.
type Error struct {
	Status int    `json:"status"` // HTTP status code
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Code   string `json:"code,omitempty"` // internal app code not used yet.
}

// NewErrorResponse creates a new formatted error response.
func NewErrorResponse(status int, title string, detail string) ErrorResponse {
	return ErrorResponse{
		Errors: []Error{
			{
				Status: status,
				Title:  title,
				Detail: detail,
			},
		},
	}
}

// WriteJSONError writes the error to json to response writer.
func WriteJSONError(w http.ResponseWriter, statusCode int, errors ErrorResponse) {
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Set the status code
	w.WriteHeader(statusCode)

	// Encode the error message as JSON
	if err := json.NewEncoder(w).Encode(errors); err != nil {
		// If encoding fails, fallback to a default error response
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// NotFoundError represents a scenario where a requested document is not found.
type NotFoundError struct {
	message string
}

// Error returns the error as a string.
func (e NotFoundError) Error() string {
	return e.message
}

// NewNotFoundError returns a new not found error.
func NewNotFoundError(message string, args ...any) NotFoundError {
	return NotFoundError{
		message: fmt.Sprintf(message, args...),
	}
}

// Is checks if err is the same as target.
func (e NotFoundError) Is(target error) bool {
	// This checks if the target error is of type NotFoundError
	_, ok := target.(NotFoundError)
	return ok
}

// BadRequestError represents a bad request error (400).
type BadRequestError struct {
	message string
}

// Error returns error as a string.
func (e BadRequestError) Error() string {
	return e.message
}

// NewBadRequestError returns a new bad request error.
func NewBadRequestError(message string, args ...any) BadRequestError {
	return BadRequestError{
		message: fmt.Sprintf(message, args...),
	}
}

// Is checks if err is the same as target.
func (e BadRequestError) Is(target error) bool {
	_, ok := target.(BadRequestError)
	return ok
}

// ForbiddenError represents a forbidden access error (403).
type ForbiddenError struct {
	message string
}

// Error returns error as a string.
func (e ForbiddenError) Error() string {
	return e.message
}

// NewForbiddenError returns a new Forbidden error.
func NewForbiddenError(message string, args ...any) ForbiddenError {
	return ForbiddenError{
		message: fmt.Sprintf(message, args...),
	}
}

// Is checks if err is the same as target.
func (e ForbiddenError) Is(target error) bool {
	_, ok := target.(ForbiddenError)
	return ok
}

// ConflictError represents a resource conflict error (409).
type ConflictError struct {
	message string
}

// Error returns the error as a string.
func (e ConflictError) Error() string {
	return e.message
}

// NewConflictError returns a new conflict error.
func NewConflictError(message string, args ...any) ConflictError {
	return ConflictError{
		message: fmt.Sprintf(message, args...),
	}
}

// Is checks if err is the same as target.
func (e ConflictError) Is(target error) bool {
	_, ok := target.(ConflictError)
	return ok
}

// func GetRequestID(r *http.Request) string {
// 	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
// 		logger.Logger.Info("Got X-Request-ID", "test", reqID)
// 		return reqID
// 	}
// 	// Fallback: generate your own if missing
// 	return uuid.NewString()
// }
