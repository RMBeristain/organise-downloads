package local_utils

import "testing"

// TestContainsHP happy path calls Contains with a slice of two elements, checking that both elements return `true`.
// To run all tests use: go test ./... (including the three dots)
func TestContainsHP(t *testing.T) {
	testSlice := []string{"one", "two", "not_tested", ""}
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{"StringNotInSlice", "hello", false},
		{"EmptyString", "", true},
		{"SingleInSlice", "one", true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Contains(testSlice, tc.input)
			if result != tc.want {
				t.Fatalf("Contains(%q) expected %v, got %v", tc.input, tc.want, result)
			}
		})
	}
}
