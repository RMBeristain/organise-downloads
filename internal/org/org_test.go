package org

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/RMBeristain/organise-downloads/internal/common"
	"github.com/RMBeristain/organise-downloads/internal/logging"
)

var excludedExtensions = []string{".DS_Store", ".localized"}

// getTestsWorkingDir returns the fully-qualified path to a directory where we can temporarily store test artifacts.
func getTestsWorkingDir() (testsWorkingDir string) {
	return filepath.Join(os.TempDir(), "organise-downloads-tests")
}

// Setup function for getFilesToMove creates some test files and returns a teardownSuite function.
func setupSuiteGetFilesToMove(tb testing.TB) (
	testFiles []fs.DirEntry,
	expectedSubDir string,
	teardownSuite func(tb testing.TB)) {

	const testFile1 string = "file1.txt"
	expectedSubDir = "txt_files"

	testWorkingDir := getTestsWorkingDir()
	tb.Logf("initialising test data in %v", testWorkingDir)

	if exists, err := common.PathExists(testWorkingDir); !exists && err == nil {
		err = os.Mkdir(testWorkingDir, 0755)
		if err != nil {
			tb.Fatalf("Unable to create test subdirectory '%v'", testWorkingDir)
		}
	}

	// add some test files
	f1 := []byte("I am File 1\n")
	err := os.WriteFile(filepath.Join(testWorkingDir, testFile1), f1, 0644)
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
		if exists, err := common.PathExists(testWorkingDir); exists && err == nil {
			files, err := os.ReadDir(testWorkingDir)
			if err != nil {
				tb.Errorf("Cannot find test subdir '%v'", testWorkingDir)
			}

			for _, this := range files {
				fqFile := filepath.Join(testWorkingDir, this.Name())
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

			workingDir := getTestsWorkingDir()
			t.Logf("working on %v", workingDir)

			// make the call we're testing
			filesToMove := GetFilesToMove(thisCase.input, &excludedExtensions)

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

	logging.InitZeroLog()
	logging.ConfiguredZerologger = logging.ConfiguredZerologger.Level(0) // Set to -1 for TraceLevel

	for _, thisCase := range table {
		t.Run(
			thisCase.name,
			func(t *testing.T) {
				t.Logf("testing %v", thisCase.input)

				workingDir := getTestsWorkingDir()
				filesToMove := GetFilesToMove(thisCase.input, &excludedExtensions)
				expectedNewDir := filepath.Join(workingDir, thisCase.expectedPath)
				filesChannel := make(chan string)

				// make the call we're testing
				go MoveFiles(workingDir, filesToMove, filesChannel)
				movedFile := <-filesChannel

				// Tests
				if len(thisCase.input) > 0 && movedFile == "" {
					t.Fatalf("expected a message, got empty string")
				} else if len(thisCase.input) == 0 && movedFile != "" {
					t.Fatalf("expected 'movedFile' to be empty string, got %v", movedFile)
				}

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
