package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	// Logger is singleton instance of logger.
	Logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	once   sync.Once
)

// SetupLogger returns the singleton logger instance
func SetupLogger() {
	once.Do(func() {
		// Default logger setup
		Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	})
}

// SetLogger allows tests to override the logger
func SetLogger(customLogger *slog.Logger) {
	Logger = customLogger
}
