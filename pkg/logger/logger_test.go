package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/flowshot-io/x/pkg/logger"
)

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := logger.New(logger.WithWriter(buf), logger.WithLogLevel("debug"))
	logger.Debug("test message")
	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected 'test message', got %v", output)
	}
}

func TestLoggerWithLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		logFn func(l logger.Logger)
	}{
		{"debug", "debug", func(l logger.Logger) { l.Debug("test message") }},
		{"info", "info", func(l logger.Logger) { l.Info("test message") }},
		{"warn", "warn", func(l logger.Logger) { l.Warn("test message") }},
		{"error", "error", func(l logger.Logger) { l.Error("test message") }},
		{"unknown", "unknown", func(l logger.Logger) { l.Info("test message") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log := logger.New(logger.WithWriter(buf), logger.WithLogLevel(tt.level))
			tt.logFn(log)
			output := buf.String()

			expectedLevel := tt.level
			if expectedLevel == "unknown" {
				expectedLevel = "info" // default level is "info"
			}

			if !strings.Contains(output, expectedLevel) {
				t.Errorf("Expected log level %v, but it was not present in output: %v", expectedLevel, output)
			}
		})
	}
}

func TestWithPretty(t *testing.T) {
	opts := &logger.Options{}
	logger.WithPretty()(opts)
	if opts.Pretty != true {
		t.Errorf("WithPretty() didn't set Pretty to true")
	}
}

func TestWithLogLevel(t *testing.T) {
	level := "debug"
	opts := &logger.Options{}
	logger.WithLogLevel(level)(opts)
	if opts.LogLevel != level {
		t.Errorf("WithLogLevel() didn't set LogLevel to %s", level)
	}
}

func TestNew(t *testing.T) {
	logger := logger.New(logger.WithPretty(), logger.WithLogLevel("debug"))
	if logger == nil {
		t.Errorf("New() returned nil")
	}
}
