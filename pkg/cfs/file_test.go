package cfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kilianpaquier/cli-sdk/pkg/cfs"
)

func TestCopyFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "file.txt")
	dest := filepath.Join(tmp, "copy.txt")

	err := os.WriteFile(src, []byte("hey file"), cfs.RwRR)
	require.NoError(t, err)

	t.Run("error_src_not_exists", func(t *testing.T) {
		// Arrange
		src := filepath.Join(tmp, "invalid.txt")

		// Act
		err := cfs.CopyFile(src, dest)

		// Assert
		assert.ErrorContains(t, err, "open")
		assert.NoFileExists(t, dest)
	})

	t.Run("error_destdir_not_exists", func(t *testing.T) {
		// Arrange
		dest := filepath.Join(tmp, "invalid", "file.txt")

		// Act
		err := cfs.CopyFile(src, dest)

		// Assert
		assert.ErrorContains(t, err, "create")
		assert.NoFileExists(t, dest)
	})

	t.Run("success", func(t *testing.T) {
		// Act
		err := cfs.CopyFile(src, dest)

		// Assert
		assert.NoError(t, err)
		assert.FileExists(t, dest)
	})

	t.Run("success_with_fs", func(t *testing.T) {
		// Act
		err := cfs.CopyFile(src, dest,
			cfs.WithFS(cfs.OS()),
			cfs.WithJoin(filepath.Join),
			cfs.WithPerm(cfs.RwRR))

		// Assert
		assert.NoError(t, err)
		assert.FileExists(t, dest)
	})
}

func TestExists(t *testing.T) {
	t.Run("false_not_exists", func(t *testing.T) {
		// Arrange
		invalid := filepath.Join(os.TempDir(), "invalid")

		// Act
		exists := cfs.Exists(invalid)

		// Assert
		assert.False(t, exists)
	})

	t.Run("true_exists", func(t *testing.T) {
		// Arrange
		srcdir := t.TempDir()
		src := filepath.Join(srcdir, "file.txt")
		file, err := os.Create(src)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		// Act
		exists := cfs.Exists(src)

		// Assert
		assert.True(t, exists)
	})
}

func TestSafeMove(t *testing.T) {
	t.Run("error_no_file", func(t *testing.T) {
		// Arrange
		dest := t.TempDir()

		// Act
		err := cfs.SafeMove(dest, dest)

		// Assert
		assert.ErrorContains(t, err, "read file")
	})

	t.Run("error_mkdir_all", func(t *testing.T) {
		// Arrange
		tmp := t.TempDir()
		src := filepath.Join(tmp, "src.txt")
		dest := filepath.Join(tmp, "subdir", "file.txt")
		require.NoError(t, os.WriteFile(src, []byte("some text"), cfs.RwRR))
		require.NoError(t, os.WriteFile(filepath.Join(tmp, "subdir"), []byte(""), cfs.RwxRxRxRx))

		// Act
		err := cfs.SafeMove(src, dest)

		// Assert
		assert.ErrorContains(t, err, "mkdir all")
	})

	t.Run("error_write_file", func(t *testing.T) {
		// Arrange
		tmp := t.TempDir()
		src := filepath.Join(tmp, "src.txt")
		dest := filepath.Join(tmp, "file.txt")
		require.NoError(t, os.WriteFile(src, []byte("some text"), cfs.RwRR))
		require.NoError(t, os.Mkdir(filepath.Join(tmp, "file.txt_"), cfs.RwxRxRxRx))

		// Act
		err := cfs.SafeMove(src, dest)

		// Assert
		assert.ErrorContains(t, err, "write file")
	})

	t.Run("error_rename", func(t *testing.T) {
		// Arrange
		tmp := t.TempDir()
		src := filepath.Join(tmp, "src.txt")
		dest := filepath.Join(tmp, "file.txt")
		require.NoError(t, os.WriteFile(src, []byte("some text"), cfs.RwRR))
		require.NoError(t, os.Mkdir(dest, cfs.RwxRxRxRx))

		// Act
		err := cfs.SafeMove(src, dest)

		// Assert
		assert.ErrorContains(t, err, "move")
	})

	t.Run("success", func(t *testing.T) {
		// Arrange
		tmp := t.TempDir()
		src := filepath.Join(tmp, "src.txt")
		dest := filepath.Join(tmp, "subdir", "file.txt")
		require.NoError(t, os.WriteFile(src, []byte("some text"), cfs.RwRR))

		// Act
		err := cfs.SafeMove(src, dest)

		// Assert
		assert.NoError(t, err)
		bytes, err := os.ReadFile(dest)
		assert.NoError(t, err)
		assert.Equal(t, []byte("some text"), bytes)
	})
}
