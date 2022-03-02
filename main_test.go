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

		for _, val := range slice {
			if val != this {
				expected = append(expected, val)
			}
		}

		slice = delSliceElement(slice, this)

		if len(slice) != len(expected) {
			t.Errorf("Slice lenght (%d) is different from expected (%d)", len(slice), len(expected))
		}
		for i, elem := range slice {
			if elem != expected[i] {
				t.Errorf("Slice element %d = %v is different from expected %v", i, elem, expected[i])
			}
		}
	}

	original := make([]string, 3)
	copy(original, slice)

	for _, this := range slice {
		var expected []string

		for _, val := range slice {
			if val != this {
				expected = append(expected, val)
			}
		}

		newSlice := delSliceElement(slice, this)

		if len(expected) != len(slice)-1 {
			t.Errorf("Expected slice lenght (%d) is the same as original (%d)", len(expected), len(slice))
		}
		if contains(newSlice, this) {
			t.Errorf("New slice %v still contains deleted value %v", newSlice, this)
		}
		if len(newSlice) != len(expected) {
			t.Errorf("New slice lenght (%d) is different from expected (%d)", len(slice), len(expected))
		}
		if len(slice) != len(original) {
			t.Errorf("Original slice lenght (%d) is different from expected (%d)", len(slice), len(original))
		}
		for i, elem := range newSlice {
			if elem != expected[i] {
				t.Errorf("New slice element %d = %v is different from expected %v", i, elem, expected[i])
			}
		}
		for i, elem := range slice {
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

func TestGetExtAndSubdir(t *testing.T) {
	testValue := []struct {
		input     string
		extension string
		subfolder string
	}{
		{"file.ext", ".ext", "ext_files"},
		{"noext", "", "_files"},
		{".onlyext", ".onlyext", "onlyext_files"},
		{".DS_Store", ".DS_Store", "DS_Store_files"},
	}

	for _, this := range testValue {
		extension, subdir := getExtAndSubdir(this.input)
		if extension != this.extension {
			t.Errorf("Returned extension '%v' doesn't match expected '%v'", extension, this.extension)
		}
		if subdir != this.subfolder {
			t.Errorf("Returned subdir '%v' doesn't match expected '%v'", subdir, this.subfolder)
		}
	}
}
