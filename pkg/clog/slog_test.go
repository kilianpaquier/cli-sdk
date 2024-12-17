package clog_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kilianpaquier/cli-sdk/pkg/clog"
)

// handler is an implementation for testing purposes to ensure that slog implementation
// for clog.Logger correctly calls the sub-functions offered by slog.
type handler struct {
	buf *bytes.Buffer
}

// Enabled implements slog.Handler.
func (*handler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (h *handler) Handle(_ context.Context, record slog.Record) error {
	_, err := h.buf.WriteString(record.Message)
	return fmt.Errorf("write string: %w", err)
}

// WithAttrs implements slog.Handler.
func (h *handler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

// WithGroup implements slog.Handler.
func (h *handler) WithGroup(string) slog.Handler {
	return h
}

var _ slog.Handler = &handler{}

func TestSlogWith(t *testing.T) {
	t.Run("debugf", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		logger := clog.Slog(slog.New(&handler{&buf}))

		// Act
		logger.Debugf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("errorf", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		logger := clog.Slog(slog.New(&handler{&buf}))

		// Act
		logger.Errorf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("infof", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		logger := clog.Slog(slog.New(&handler{&buf}))

		// Act
		logger.Infof("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("warnf", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		logger := clog.Slog(slog.New(&handler{&buf}))

		// Act
		logger.Warnf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})
}
