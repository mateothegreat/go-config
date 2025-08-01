package util

import (
	"log"
	"path/filepath"
	"runtime"
)

func GetCallerDir() string {
	// Get the directory of the current file and construct path to config.yaml
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Failed to get current file path")
	}

	return filepath.Dir(currentFile)
}
