package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/RMBeristain/organise-downloads/local_utils"
)

// Test function 'contains'
func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	// Happy
	for _, this := range slice {
		result := local_utils.Contains(slice, this)

		if !result {
			t.Errorf("Expected 'true' for '%v' but got 'false'", this)
		}
	}

	// Errors
	if local_utils.Contains(slice, "X") != false {
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
		if local_utils.Contains(newSlice, this) {
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

// Setup function for getFilesToMove
func setupSuiteGetFilesToMove(tb testing.TB) func(tb testing.TB) {
	//Create test subdir organise-downloads-tests
	// we assume having permissions to the path since that's where we work

	var testDir string = "Downloads"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		tb.Fatalf("Unable to determine user home dir - %v", err)
	}

	if this_os := runtime.GOOS; this_os == "linux" {
		homeDir = "/tmp"
		testDir = ""
	}

	testPath := filepath.Join(homeDir, testDir, "organise-downloads-tests")

	if exists, err := pathExists(testPath); !exists && err == nil {
		err = os.Mkdir(testPath, 0755)
		if err != nil {
			tb.Fatalf("Unable to create test subdirectory '%v'", testPath)
		}

		// add some test files
		f1 := []byte("File 1\n")
		err := os.WriteFile(path.Join(testPath, "file1.txt"), f1, 0644)
		if err != nil {
			tb.Errorf("Cannot add test file %v", testPath)
		}
	}

	// Return a function to tear down the suite
	return func(tb testing.TB) {
		if exists, err := pathExists(testPath); exists && err == nil {
			files, err := ioutil.ReadDir(testPath)
			if err != nil {
				tb.Errorf("Cannot find test subdir '%v'", testPath)
			}

			for _, this := range files {
				fqFile := path.Join(testPath, this.Name())
				err = os.RemoveAll(fqFile)
				if err != nil {
					tb.Errorf("Unable to delete test file '%v'", fqFile)
				}
			}

			err = os.Remove(testPath)
			if err != nil {
				tb.Errorf("Unable to delete test subdir %v", testPath)
			}
		}
	}
}

func TestGetFilesToMove(t *testing.T) {
	teardownSuite := setupSuiteGetFilesToMove(t)
	defer teardownSuite(t)

	table := []struct {
		name     string
		input    string
		expected string
	}{
		{"one", "Hi mom!", "Hi mom!"},
	}

	for _, this := range table {
		t.Run(this.name, func(t *testing.T) {
			result := bla(this.input)
			if result != this.expected {
				t.Errorf("expected %v, got %v", this.expected, result)
			}
		})
	}
}
