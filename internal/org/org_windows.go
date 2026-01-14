//go:build windows

package org

import "os"

// isFileInUse checks if a file is locked by another process by attempting to open it.
func isFileInUse(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return true
	}
	file.Close()
	return false
}
