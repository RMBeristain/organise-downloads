package main

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/RMBeristain/organise-downloads/local_utils"
)

const defaultWorkingDir string = "Downloads"

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

// getTestsWorkingDir returns the fully-qualified path to a directory where we can temporarily store test artifacts.
// By default, we will create a subfolder within ~/Downloads because we assume we'll have write permission there.
func getTestsWorkingDir(tb testing.TB) (testsWorkingDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		tb.Fatalf("Unable to determine user home dir - %v", err)
	}
	return filepath.Join(homeDir, defaultWorkingDir, "organise-downloads-tests")
}

// Setup function for getFilesToMove creates some test files and returns a teardownSuite function.
func setupSuiteGetFilesToMove(tb testing.TB) (
	testFiles []fs.DirEntry,
	expectedSubDir string,
	teardownSuite func(tb testing.TB)) {

	const testFile1 string = "file1.txt"
	expectedSubDir = "txt_files"

	testWorkingDir := getTestsWorkingDir(tb)
	tb.Logf("initialising test data in %v", testWorkingDir)

	if exists, err := pathExists(testWorkingDir); !exists && err == nil {
		err = os.Mkdir(testWorkingDir, 0755)
		if err != nil {
			tb.Fatalf("Unable to create test subdirectory '%v'", testWorkingDir)
		}
	}

	// add some test files
	f1 := []byte("I am File 1\n")
	err := os.WriteFile(path.Join(testWorkingDir, testFile1), f1, 0644)
	if err != nil {
		tb.Errorf("Cannot add test file %v at %v", testFile1, testWorkingDir)
	}
	tb.Logf("wrote test file %v/%v", testWorkingDir, testFile1)

	// get the list of test files
	testFiles, err = os.ReadDir(testWorkingDir)
	if err != nil {
		tb.Errorf("Cannot read test dir %v", testWorkingDir)
	}
	tb.Logf("will test over (%v) file(s) %v", len(testFiles), testFiles)

	// Return a function to tear down the suite
	return testFiles, expectedSubDir, func(tb testing.TB) {
		if exists, err := pathExists(testWorkingDir); exists && err == nil {
			files, err := os.ReadDir(testWorkingDir)
			if err != nil {
				tb.Errorf("Cannot find test subdir '%v'", testWorkingDir)
			}

			for _, this := range files {
				fqFile := path.Join(testWorkingDir, this.Name())
				err = os.RemoveAll(fqFile)
				if err != nil {
					tb.Errorf("Unable to delete test file '%v'", fqFile)
				}
			}

			err = os.Remove(testWorkingDir)
			if err != nil {
				tb.Errorf("Unable to delete test subdir %v", testWorkingDir)
			}
		}
	}
}

func TestGetFilesToMove(t *testing.T) {
	testFiles, testsWorkingDir, teardownSuite := setupSuiteGetFilesToMove(t)
	defer teardownSuite(t)

	var emptyDirSlice []fs.DirEntry
	table := []struct {
		name         string
		input        []fs.DirEntry
		expected     []fs.DirEntry
		expectedPath string
	}{
		{
			"one file", testFiles, testFiles, testsWorkingDir,
		},
		{
			"no files", emptyDirSlice, emptyDirSlice, "some_path_that_shouldnt_exist",
		},
	}
	for _, thisCase := range table {
		t.Run(thisCase.name, func(t *testing.T) {
			t.Logf("testing %v", thisCase.input)

			workingDir := *getCurrentUserDownloadPath()

			// make the call we're testing
			t.Logf("working on %v", workingDir)
			filesToMove := getFilesToMove(thisCase.input)

			// Tests
			if len(filesToMove) == 0 {
				if thisCase.name == "no files" {
					t.Skipf("OK - skipping dir without files")
				}
				t.Errorf("expected to find files to move, got %v", filesToMove)
			}

			matches := make(map[string]int) // number of files with this name we expect to find
			for _, file := range thisCase.expected {
				matches[file.Name()]++
			}
			for destination, files := range filesToMove {
				if destination != thisCase.expectedPath {
					t.Errorf("expected %v to match %v", destination, thisCase.expectedPath)
				}
				for _, file := range files {
					if count, ok := matches[file]; !ok {
						t.Errorf("expected count of %v to be %v, got unknown file", file, count)
					} else if count == 0 {
						t.Errorf("expected count of %v to be %v, got not matches", file, count)

					}
					matches[file]--
				}
			}
		})
	}
}

func TestMoveFile(t *testing.T) {
	testFiles, testsWorkingDir, teardownSuite := setupSuiteGetFilesToMove(t)
	defer teardownSuite(t)

	var emptyDirSlice []fs.DirEntry
	table := []struct {
		name         string
		input        []fs.DirEntry
		expectedPath string
	}{
		{
			"one file", testFiles, testsWorkingDir,
		},
		{
			"no files", emptyDirSlice, testsWorkingDir,
		},
	}

	initLoggingToFile()

	for _, thisCase := range table {
		t.Run(
			thisCase.name,
			func(t *testing.T) {
				t.Logf("testing %v", thisCase.input)

				workingDir := getTestsWorkingDir(t)
				filesToMove := getFilesToMove(thisCase.input)
				expectedNewDir := filepath.Join(workingDir, thisCase.expectedPath)

				// make the call we're testing
				moveFiles(workingDir, filesToMove)

				// Tests
				files, err := os.ReadDir(expectedNewDir)
				if err != nil {
					t.Fatalf("expected to read files from %v, got %v", expectedNewDir, err.Error())
				}

				matches := make(map[string]int) // number of files with this name we expect to find
				for _, file := range files {
					matches[file.Name()]++
				}
				for _, files := range filesToMove {
					for _, expectedFile := range files {
						if count, ok := matches[expectedFile]; !ok {
							t.Errorf("expected count of %v to be %v, got unknown file", expectedFile, count)
						} else if count == 0 {
							t.Errorf("expected count of %v to be %v, got not matches", expectedFile, count)
						}
						matches[expectedFile]--
					}
				}
			},
		)
	}

}
