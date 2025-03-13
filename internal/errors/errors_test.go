package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		errorResponse      ErrorResponse
		inputStatusCode    int
		expectedStatusCode int
		expectedErrorTitle string
	}{
		{
			name: "basic not found error",
			errorResponse: NewErrorResponse(
				http.StatusNotFound,
				"Resource Not Found",
				"The requested resource could not be found.",
			),
			inputStatusCode:    http.StatusNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedErrorTitle: "Resource Not Found",
		},
		{
			name: "bad request error",
			errorResponse: NewErrorResponse(
				http.StatusBadRequest,
				"Invalid Input",
				"Some fields are missing.",
			),
			inputStatusCode:    http.StatusBadRequest,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorTitle: "Invalid Input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()

			WriteJSONError(rr, tt.inputStatusCode, tt.errorResponse)

			// Validate the response
			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var parsedResp ErrorResponse
			err := json.NewDecoder(rr.Body).Decode(&parsedResp)
			assert.NoError(t, err)
			assert.Len(t, parsedResp.Errors, 1)

			errorEntry := parsedResp.Errors[0]
			assert.Equal(t, tt.expectedStatusCode, errorEntry.Status)
			assert.Equal(t, tt.expectedErrorTitle, errorEntry.Title)
		})
	}
}

func TestErrorTypes_IsMethod(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		err           error
		targetErr     error
		expectedMatch bool
	}{
		{
			name:          "NotFoundError match",
			err:           NewNotFoundError("not found"),
			targetErr:     NotFoundError{},
			expectedMatch: true,
		},
		{
			name:          "BadRequestError match",
			err:           NewBadRequestError("bad request"),
			targetErr:     BadRequestError{},
			expectedMatch: true,
		},
		{
			name:          "ForbiddenError match",
			err:           NewForbiddenError("forbidden"),
			targetErr:     ForbiddenError{},
			expectedMatch: true,
		},
		{
			name:          "ConflictError match",
			err:           NewConflictError("conflict"),
			targetErr:     ConflictError{},
			expectedMatch: true,
		},
		{
			name:          "No match with random error",
			err:           NewBadRequestError("bad request"),
			targetErr:     errors.New("random error"),
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			match := errors.Is(tt.err, tt.targetErr)
			assert.Equal(t, tt.expectedMatch, match)
		})
	}
}
