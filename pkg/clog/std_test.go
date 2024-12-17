package clog_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kilianpaquier/cli-sdk/pkg/clog"
)

func TestStd(t *testing.T) {
	t.Run("debugf", func(t *testing.T) {
		// Arrange
		logger := clog.Std(log.Default())
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		logger.Debugf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("errorf", func(t *testing.T) {
		// Arrange
		logger := clog.Std(log.Default())
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		logger.Errorf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("infof", func(t *testing.T) {
		// Arrange
		logger := clog.Std(log.Default())
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		logger.Infof("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})

	t.Run("warnf", func(t *testing.T) {
		// Arrange
		logger := clog.Std(log.Default())
		var buf bytes.Buffer
		log.SetOutput(&buf)

		// Act
		logger.Warnf("some message")

		// Assert
		assert.Contains(t, buf.String(), "some message")
	})
}
