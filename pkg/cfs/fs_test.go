package cfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kilianpaquier/cli-sdk/pkg/cfs"
)

func TestOS(t *testing.T) {
	tmp := t.TempDir()

	name := filepath.Join(tmp, "hey.txt")
	err := os.WriteFile(name, []byte("hey !"), cfs.RwRR)
	require.NoError(t, err)

	fsys := cfs.OS()

	t.Run("success_open", func(t *testing.T) {
		// Act
		file, err := fsys.Open(name)
		require.NoError(t, err)
		defer file.Close()

		// Assert
		assert.NotNil(t, file)
	})

	t.Run("success_read_dir", func(t *testing.T) {
		// Act
		entries, err := fsys.ReadDir(tmp)

		// Assert
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, "hey.txt", entries[0].Name())
	})

	t.Run("success_read_file", func(t *testing.T) {
		// Act
		bytes, err := fsys.ReadFile(name)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "hey !", string(bytes))
	})
}
