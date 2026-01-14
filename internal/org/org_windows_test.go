//go:build windows

package org

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFileInUse_Windows(t *testing.T) {
	t.Run("File is locked", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "locked.txt")

		// Create and lock the file by opening it for writing.
		// On Windows, this prevents other processes from opening it.
		lockedFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			t.Fatalf("Failed to create and lock file: %v", err)
		}
		defer lockedFile.Close()

		if !isFileInUse(filePath) {
			t.Error("Expected isFileInUse to be true for a locked file")
		}
	})

	t.Run("File does not exist", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "nonexistent.txt")

		if !isFileInUse(filePath) {
			t.Error("Expected isFileInUse to be true for a non-existent file")
		}
	})
}
