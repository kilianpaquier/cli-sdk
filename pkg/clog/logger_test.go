package clog_test

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kilianpaquier/cli-sdk/pkg/clog"
)

func TestGetLogger(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		logger := clog.Std(log.Default())
		ctx := context.WithValue(context.Background(), clog.LoggerKey, logger)

		// Act
		actual := clog.GetLogger(ctx)

		// Assert
		assert.Equal(t, logger, actual)
	})
}
