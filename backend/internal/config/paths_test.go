package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/config"
)

func TestCheckPaths(t *testing.T) {
	t.Run("happy path - creates media folder", func(t *testing.T) {
		// given
		root := t.TempDir()
		cfg := config.Config{DBPath: filepath.Join(root, "halo.db"), MediaDir: filepath.Join(root, "media")}

		// when
		err := config.CheckPaths(cfg)

		// then
		require.NoError(t, err)
		info, err := os.Stat(cfg.MediaDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
		entries, err := os.ReadDir(root)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})

	t.Run("error - database folder not writable", func(t *testing.T) {
		// given
		if os.Geteuid() == 0 {
			t.Skip("root ignores folder permissions")
		}
		root := t.TempDir()
		dbDir := filepath.Join(root, "db")
		require.NoError(t, os.Mkdir(dbDir, 0o500))
		t.Cleanup(func() { _ = os.Chmod(dbDir, 0o700) })
		cfg := config.Config{DBPath: filepath.Join(dbDir, "halo.db"), MediaDir: filepath.Join(root, "media")}

		// when
		err := config.CheckPaths(cfg)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checking")
	})

	t.Run("error - media folder not writable", func(t *testing.T) {
		// given
		if os.Geteuid() == 0 {
			t.Skip("root ignores folder permissions")
		}
		root := t.TempDir()
		mediaDir := filepath.Join(root, "media")
		require.NoError(t, os.Mkdir(mediaDir, 0o500))
		t.Cleanup(func() { _ = os.Chmod(mediaDir, 0o700) })
		cfg := config.Config{DBPath: filepath.Join(root, "halo.db"), MediaDir: mediaDir}

		// when
		err := config.CheckPaths(cfg)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checking "+mediaDir)
	})

	t.Run("error - database folder missing", func(t *testing.T) {
		// given
		root := t.TempDir()
		cfg := config.Config{
			DBPath:   filepath.Join(root, "missing", "halo.db"),
			MediaDir: filepath.Join(root, "media"),
		}

		// when
		err := config.CheckPaths(cfg)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checking "+filepath.Join(root, "missing"))
	})

	t.Run("error - media folder cannot be created", func(t *testing.T) {
		// given
		root := t.TempDir()
		blocker := filepath.Join(root, "file")
		require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
		cfg := config.Config{DBPath: filepath.Join(root, "halo.db"), MediaDir: filepath.Join(blocker, "media")}

		// when
		err := config.CheckPaths(cfg)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating media folder")
	})
}
