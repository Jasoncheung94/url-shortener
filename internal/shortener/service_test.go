package shortener

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/mocks"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"go.uber.org/mock/gomock"
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

func TestShortenerService_SaveURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		data          model.URL
		mockBehavior  func(m *mocks.MockURL)
		expected      model.URL
		expectedError error
	}{
		{
			name: "Success",
			data: model.URL{OriginalURL: "https://example.com"},
			mockBehavior: func(m *mocks.MockURL) {
				m.EXPECT().IncrementCounter().Return(uint64(100), nil)
				m.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: model.URL{ShortURL: "1C"},
		},
		{
			name:          "Invalid URL",
			data:          model.URL{OriginalURL: "invalid-url"},
			mockBehavior:  func(m *mocks.MockURL) {}, // No calls expected
			expectedError: errors.New("invalid url"),
		},
		{
			name: "Repository Error",
			data: model.URL{OriginalURL: "https://example.com"},
			mockBehavior: func(m *mocks.MockURL) {
				m.EXPECT().IncrementCounter().Return(uint64(1), nil)
				m.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			expectedError: errors.New("shortener/service: failed to create url: repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockURL(ctrl)
			tt.mockBehavior(mockRepo)

			service := NewService(mockRepo)
			shortURL, err := service.SaveURL(context.Background(), &tt.data)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ShortURL, shortURL)
			}
		})
	}
}

func TestShortenerService_GetURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		input         string
		mockBehavior  func(m *mocks.MockURL)
		expected      model.URL
		expectedError error
	}{
		{
			name:  "Success",
			input: "abc123",
			mockBehavior: func(m *mocks.MockURL) {
				m.EXPECT().GetURL(gomock.Any(), "abc123").Return(&model.URL{
					OriginalURL: "https://example.com",
				}, nil)
			},
			expected: model.URL{OriginalURL: "https://example.com"},
		},
		{
			name:          "Invalid Short URL",
			input:         "!!invalid!!",
			mockBehavior:  func(m *mocks.MockURL) {}, // No call expected
			expectedError: errors.New("invalid url"),
		},
		{
			name:  "Self-referencing Redirect (Invalid)",
			input: "http://localhost:8080/Lu8S545",
			mockBehavior: func(m *mocks.MockURL) {
				// nothing expected here since validation should fail before call
			},
			expectedError: errors.New("invalid url"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockURL(ctrl)
			tt.mockBehavior(mockRepo)

			service := NewService(mockRepo)
			result, err := service.GetURL(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.OriginalURL, result.OriginalURL)
			}
		})
	}
}
