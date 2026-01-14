//go:build !windows

package org

// isFileInUse returns false on non-Windows systems as file locking is advisory.
// We skip this check to avoid skipping files due to permission errors (which os.Rename might handle).
func isFileInUse(_ string) bool {
	return false
}
