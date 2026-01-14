package org

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RMBeristain/organise-downloads/internal/common"
	"github.com/RMBeristain/organise-downloads/internal/logging"
	"github.com/rs/zerolog"
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
			filesToMove := GetFilesToMove(thisCase.input, excludedExtensions)

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
				filesToMove := GetFilesToMove(thisCase.input, excludedExtensions)
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

// mockDirEntry implements fs.DirEntry for testing purposes.
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() fs.FileMode          { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestGetFilesToMove_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    []fs.DirEntry
		excluded []string
		validate func(t *testing.T, targets map[string][]string)
	}{
		{
			name: "Directory entry",
			input: []fs.DirEntry{
				mockDirEntry{name: "subdir", isDir: true},
			},
			excluded: []string{},
			validate: func(t *testing.T, targets map[string][]string) {
				if _, ok := targets["subdir"]; !ok {
					t.Error("Expected directory 'subdir' to be in targets")
				}
				if len(targets["subdir"]) != 0 {
					t.Errorf("Expected empty slice for directory, got %v", targets["subdir"])
				}
			},
		},
		{
			name: "Excluded file",
			input: []fs.DirEntry{
				mockDirEntry{name: "file.tmp", isDir: false},
			},
			excluded: []string{".tmp"},
			validate: func(t *testing.T, targets map[string][]string) {
				if len(targets) != 0 {
					t.Errorf("Expected no targets for excluded file, got %v", targets)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := GetFilesToMove(tt.input, tt.excluded)
			tt.validate(t, targets)
		})
	}
}

func TestMoveFiles_EdgeCases(t *testing.T) {
	logging.InitZeroLog()
	logging.ConfiguredZerologger = logging.ConfiguredZerologger.Level(zerolog.Disabled)

	t.Run("Destination file already exists", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "org_test_conflict")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		fileName := "conflict.txt"
		subDir := "txt_files"

		// Create source file
		if err := os.WriteFile(filepath.Join(tmpDir, fileName), []byte("source"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create destination dir and file
		destDir := filepath.Join(tmpDir, subDir)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(destDir, fileName), []byte("dest"), 0644); err != nil {
			t.Fatal(err)
		}

		filesToMove := map[string][]string{
			subDir: {fileName},
		}
		filesChannel := make(chan string, 1)

		MoveFiles(tmpDir, filesToMove, filesChannel)

		// Expectation: Channel receives the path, but file is not moved.
		select {
		case <-filesChannel:
			// OK
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected file path in channel even if skipped")
		}

		// Verify source file still exists
		if _, err := os.Stat(filepath.Join(tmpDir, fileName)); os.IsNotExist(err) {
			t.Error("Source file should still exist when destination conflicts")
		}
	})

	t.Run("Source file missing (Rename error)", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "org_test_missing")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		filesToMove := map[string][]string{
			"any_dir": {"missing.txt"},
		}
		filesChannel := make(chan string, 1)

		MoveFiles(tmpDir, filesToMove, filesChannel)

		// Expectation: Channel receives nothing (rename fails -> continue).
		select {
		case msg := <-filesChannel:
			if msg != "" {
				t.Errorf("Expected no message in channel for missing file, got %s", msg)
			}
		case <-time.After(100 * time.Millisecond):
			// OK
		}
	})
}

func TestIsFileInUse(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Happy path: file exists and is not locked.
	if isFileInUse(file) {
		t.Errorf("Expected file to be reported as not in use")
	}
}
