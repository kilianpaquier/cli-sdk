package clog_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/kilianpaquier/cli-sdk/pkg/clog"

	"github.com/stretchr/testify/assert"
)

func TestColorHook(t *testing.T) {
	t.Run("debugf", func(t *testing.T) {
		// Arrange
		hook := clog.NewHook(clog.Std(log.Default()), clog.ColorHook)
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		hook.Debugf("some message")

		// Assert
		assert.Contains(t, buf.String(), clog.Gray+"some message"+clog.Reset)
	})

	t.Run("errorf", func(t *testing.T) {
		// Arrange
		hook := clog.NewHook(clog.Std(log.Default()), clog.ColorHook)
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		hook.Errorf("some message")

		// Assert
		assert.Contains(t, buf.String(), clog.Red+"some message"+clog.Reset)
	})

	t.Run("infof", func(t *testing.T) {
		// Arrange
		hook := clog.NewHook(clog.Std(log.Default()), clog.ColorHook)
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		hook.Infof("some message")

		// Assert
		assert.Contains(t, buf.String(), clog.Cyan+"some message"+clog.Reset)
	})

	t.Run("warnf", func(t *testing.T) {
		// Arrange
		hook := clog.NewHook(clog.Std(log.Default()), clog.ColorHook)
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		hook.Warnf("some message")

		// Assert
		assert.Contains(t, buf.String(), clog.Yellow+"some message"+clog.Reset)
	})
}
