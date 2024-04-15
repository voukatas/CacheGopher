package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// interface to decouple logging from logger
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type SlogWrapper struct {
	logger *slog.Logger
}

func (s *SlogWrapper) Debug(msg string) { s.logger.Debug(msg) }
func (s *SlogWrapper) Info(msg string)  { s.logger.Info(msg) }
func (s *SlogWrapper) Warn(msg string)  { s.logger.Warn(msg) }
func (s *SlogWrapper) Error(msg string) { s.logger.Error(msg) }

func SetupLogger(logFileName, level string, prod bool) (Logger, func()) {

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

	if prod {
		handler = slog.NewTextHandler(logOutput, opts)
	}

	logger := slog.New(handler)

	slogWrapper := &SlogWrapper{logger}

	// constructor-cleanup idiom
	return slogWrapper, func() {
		fmt.Println("closing log file")
		logOutput.Close()
	}
}

func SetupDebugLogger() Logger {

	logLevel := slog.LevelDebug

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	logger := slog.New(handler)
	slogWrapper := &SlogWrapper{logger}

	return slogWrapper
}
