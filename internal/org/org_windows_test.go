//go:build windows

package org

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestIsFileInUse_Windows(t *testing.T) {
	t.Run("File is locked", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "locked.txt")

		// Create and lock the file exclusively.
		// os.OpenFile uses shared mode, so we use syscall to enforce exclusive access.
		pathPtr, err := syscall.UTF16PtrFromString(filePath)
		if err != nil {
			t.Fatalf("Failed to convert path: %v", err)
		}
		handle, err := syscall.CreateFile(
			pathPtr,
			syscall.GENERIC_READ|syscall.GENERIC_WRITE,
			0, // 0 = Exclusive access (no sharing)
			nil,
			syscall.CREATE_ALWAYS,
			syscall.FILE_ATTRIBUTE_NORMAL,
			0,
		)
		if err != nil {
			t.Fatalf("Failed to create and lock file: %v", err)
		}
		defer syscall.CloseHandle(handle)

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
