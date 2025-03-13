package shortener

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/mocks"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
)

// This test runs before everything else in the package.
// Required to setup the logger for all tests instead of manual setup per test.
func TestMain(m *testing.M) {
	// Set up test logger
	logger.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil)) // io.Discard

	// Run all tests
	code := m.Run()

	// Exit with the proper code
	os.Exit(code)
}

func TestShortenerService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockURL(ctrl)
	service := NewService(mockRepo) // Logger is nil for testing

	// Table-driven test cases
	tests := []struct {
		name          string
		method        string
		input         string
		data          model.URL
		mockBehavior  func()
		expected      model.URL
		expectedError error
	}{
		// --- SaveURL cases ---
		{
			name:   "SaveURL - Success",
			method: "SaveURL",
			data: model.URL{
				OriginalURL: "https://example.com",
			},
			mockBehavior: func() {
				mockRepo.EXPECT().IncrementCounter().Return(uint64(100), nil)
				mockRepo.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: model.URL{ShortURL: "1C"},
		},
		{
			name:          "SaveURL - Invalid URL",
			method:        "SaveURL",
			input:         "invalid-url",
			mockBehavior:  func() {}, // No repo call expected
			expectedError: errors.New("invalid url"),
		},
		{
			name:   "SaveURL - Repository Error",
			method: "SaveURL",
			data: model.URL{
				OriginalURL: "https://example1.com",
			},
			mockBehavior: func() {
				mockRepo.EXPECT().IncrementCounter().Return(uint64(1), nil)
				mockRepo.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			expectedError: errors.New("shortener/service: failed to create url: repository error"),
		},

		// --- GetURL cases ---
		{
			name:   "GetURL - Success",
			method: "GetURL",
			input:  "abc123",
			mockBehavior: func() {
				mockRepo.EXPECT().GetURL(gomock.Any(), "abc123").Return(&model.URL{
					OriginalURL: "https://example.com",
				}, nil)

			},
			expected: model.URL{OriginalURL: "https://example.com"},
		},
		{
			name:          "GetURL - Invalid Short URL",
			method:        "GetURL",
			input:         "!!invalid!!",
			mockBehavior:  func() {}, // No repo call expected
			expectedError: errors.New("invalid url"),
		},
		{
			name:   "GetURL - Not Found",
			method: "GetURL",
			input:  "http://localhost:8080/Lu8S545",
			mockBehavior: func() {
				//nolint:lll
				// mockRepo.EXPECT().GetURL(gomock.Any(), "http://localhost:8080/Lu8S545").Return("", errors.New("failed to get url"))
			},
			expectedError: errors.New("invalid url"),
		},
	}

	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.mockBehavior() // Set up mocks

			result := &model.URL{}
			var err error

			// Call the appropriate method
			switch tc.method {
			case "SaveURL":
				result.ShortURL, err = service.SaveURL(context.Background(), &tc.data)
			case "GetURL":
				result, err = service.GetURL(context.Background(), tc.input)
			}

			// Verify
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error(),
					"Unexpected output:\nExpected: %q\nGot: %q", tc.expectedError.Error(), err)
				// assert.EqualValues()
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.ShortURL, result.ShortURL,
					"Unexpected output:\nExpected: %+v\nGot: %+v", tc.expected.ShortURL, result.ShortURL)
				assert.Equal(t, tc.expected.OriginalURL, result.OriginalURL,
					"Unexpected output:\nExpected: %+v\nGot: %+v", tc.expected.OriginalURL, result.OriginalURL)
			}
		})
	}
}
