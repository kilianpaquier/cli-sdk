package logger_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kilianpaquier/cli-sdk/pkg/logger"
)

func TestStd(t *testing.T) {
	t.Run("infof", func(t *testing.T) {
		// Arrange
		std := logger.Std()
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		std.Infof("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("warnf", func(t *testing.T) {
		// Arrange
		std := logger.Std()
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		std.Warnf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})
}

func TestStdWith(t *testing.T) {
	t.Run("infof", func(t *testing.T) {
		// Arrange
		std := logger.StdWith(log.Default())
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		std.Infof("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("warnf", func(t *testing.T) {
		// Arrange
		std := logger.StdWith(log.Default())
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		std.Warnf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})
}
