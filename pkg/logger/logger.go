package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func SetupLogger(logFileName, level string) (*slog.Logger, func()) {

	var logLevel slog.Level

	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logOutput, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("failed to open log file: ", err)
	}

	mw := io.MultiWriter(logOutput, os.Stdout)
	handler := slog.NewTextHandler(mw, opts)

	logger := slog.New(handler)

	// constructor-cleanup idiom
	return logger, func() {
		fmt.Println("closing log file")
		logOutput.Close()
	}
}

func SetupTestLogger() *slog.Logger {

	logLevel := slog.LevelDebug

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	logger := slog.New(handler)

	return logger
}
