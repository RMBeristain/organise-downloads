package main

import (
	"testing"
)

// Test private function 'contains'
func TestPrivateContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	// Happy
	for _, this := range slice {
		result := contains(slice, this)

		if !result {
			t.Errorf("Expected 'true' for '%v' but got 'false'", this)
		}
	}

	// Errors
	if contains(slice, "X") != false {
		t.Errorf("Expected 'false' but got 'true'")
	}
}

func TestPrivatePathExists(t *testing.T) {
	goodPath := "."
	badPath := "/I_should_not_exist"

	if exists, err := pathExists(goodPath); !exists {
		t.Errorf("Cannot find '.'")
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if exists, _ := pathExists(badPath); exists {
		t.Errorf("Found '%v' but expected it to not exist.", badPath)
	}
}

func TestPrivateDelSliceElement(t *testing.T) {
	slice := []string{"uno", "dos", "tres", "cuatro", "cinco"}

	// Happy
	for _, this := range slice {
		var expected []string
		testSlice := make([]string, len(slice))

		copy(testSlice, slice)

		for _, val := range testSlice {
			if val != this {
				expected = append(expected, val)
			}
		}

		testSlice = delSliceElement(testSlice, this)

		if len(testSlice) != len(expected) {
			t.Errorf("Slice lenght (%d) is different from expected (%d)", len(testSlice), len(expected))
		}
		for i, elem := range testSlice {
			if elem != expected[i] {
				t.Errorf("Slice element %d = %v is different from expected %v", i, elem, expected[i])
			}
		}
	}

	for _, this := range slice {
		var expected []string
		testSlice := make([]string, len(slice))

		copy(testSlice, slice)

		for _, val := range testSlice {
			if val != this {
				expected = append(expected, val)
			}
		}

		newSlice := delSliceElement(testSlice, this)

		if len(expected) != len(slice)-1 {
			t.Errorf("Expected slice lenght (%d) is the same as original (%d)", len(expected), len(slice))
		}
		if contains(newSlice, this) {
			t.Errorf("New slice %v still contains deleted value %v", newSlice, this)
		}
		if len(newSlice) != len(expected) {
			t.Errorf("New slice lenght (%d) is different from expected (%d)", len(testSlice), len(expected))
		}
		if len(testSlice) != len(slice) {
			t.Errorf("Original slice lenght (%d) is different from expected (%d)", len(testSlice), len(slice))
		}
		for i, elem := range newSlice {
			if elem != expected[i] {
				t.Errorf("New slice element %d = %v is different from expected %v", i, elem, expected[i])
			}
		}
		for i, elem := range testSlice {
			if elem != slice[i] {
				t.Errorf("Original slice element %d = %v is different from expected %v", i, elem, slice[i])
			}
		}
	}

	// Error
	newSlice := delSliceElement(slice, "not there")
	for i, elem := range newSlice {
		if elem != slice[i] {
			t.Errorf("New slice element %d = %v is different from expected %v", i, elem, slice[i])
		}
	}
}
