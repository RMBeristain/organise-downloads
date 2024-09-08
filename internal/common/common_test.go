package common

import (
	"testing"
)

func TestPathExists(t *testing.T) {
	goodPath := "."
	badPath := "/I_should_not_exist"

	if exists, err := PathExists(goodPath); !exists {
		t.Errorf("Unable to find '.'")
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if exists, _ := PathExists(badPath); exists {
		t.Errorf("Expected '%v' to not exist, got exists=%v", badPath, exists)
	}
}

func TestGetExtAndSubdir(t *testing.T) {
	testCases := []struct {
		input     string
		extension string
		subfolder string
	}{
		{"file.ext", ".ext", "ext_files"},
		{"noext", "", "_files"},
		{".onlyext", ".onlyext", "onlyext_files"},
		{".DS_Store", ".DS_Store", "DS_Store_files"},
	}

	for _, this := range testCases {
		extension, subdir := GetExtAndSubdir(this.input)
		if extension != this.extension {
			t.Errorf("Returned extension '%v' doesn't match expected '%v'", extension, this.extension)
		}
		if subdir != this.subfolder {
			t.Errorf("Returned subdir '%v' doesn't match expected '%v'", subdir, this.subfolder)
		}
	}
}
