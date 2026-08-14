package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	configFile string
	backupFile string
	doBackup   bool
	validate   bool
}

func ParseFlags() Config {
	cf := flag.String("config_file", "", "determines the path to a file that holds config")
	bf := flag.String("backup_file", "", "determines the path at which backup_file should be created")
	backup := flag.Bool("backup", false, "switches functionality to only create a backup")
	validate := flag.Bool("validate", false, "switches functionality to only validate the config")
	flag.Parse()
	return Config{configFile: *cf, backupFile: *bf, doBackup: *backup, validate: *validate}
}

func ValidateFilePath(path string) error {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("path does not exist: %s", cleanPath)
		}
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("permission denied: %s", cleanPath)
		}
		return fmt.Errorf("cannot access path: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory: %s", cleanPath)
	}

	return nil
}
