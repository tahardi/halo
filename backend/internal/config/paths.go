package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const mediaDirMode = 0o700

func CheckPaths(cfg Config) error {
	if err := os.MkdirAll(cfg.MediaDir, mediaDirMode); err != nil {
		return fmt.Errorf("creating media folder: %w", err)
	}
	for _, dir := range []string{cfg.MediaDir, filepath.Dir(cfg.DBPath)} {
		if err := checkWritable(dir); err != nil {
			return err
		}
	}
	return nil
}

func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".halo-write-check-*")
	if err != nil {
		return fmt.Errorf("checking %s is writable: %w", dir, err)
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing write check file: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("removing write check file: %w", err)
	}
	return nil
}
