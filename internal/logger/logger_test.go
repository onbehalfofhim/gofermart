package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	l := &Logger{log: slog.New(handler)}

	l.Info("test info message", "key", "value")

	output := buf.String()

	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, `"level":"INFO"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	l := &Logger{log: slog.New(handler)}

	l.Error("test error message", "error", "something failed")

	output := buf.String()

	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, `"level":"ERROR"`)
	assert.Contains(t, output, `"error":"something failed"`)
}
