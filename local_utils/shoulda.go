// Utilities that shoulda been shipped with go :P

package local_utils

// Check whether 'slice' contains 'element'.
func Contains(slice []string, element string) bool {
	for _, this := range slice {
		if this == element {
			return true
		}
	}
	return false
}
