package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathExists(t *testing.T) {
	goodPath := "."
	badPath := filepath.Join(t.TempDir(), "should-not-exist")

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

func TestGetCurrentUserDownloadPath(t *testing.T) {
	// Since we cannot mock user.Current or os.UserHomeDir without refactoring,
	// we test the happy path on the running system.
	t.Run("Happy Path", func(t *testing.T) {
		defaultSrc := "Downloads"
		path, err := GetCurrentUserDownloadPath(defaultSrc)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if path == "" {
			t.Error("Expected a path, got empty string")
		}
		if filepath.Base(path) != defaultSrc {
			t.Errorf("Expected path ending in %s, got %s", defaultSrc, path)
		}
	})
}

func TestCreateDirIfNotExists(t *testing.T) {
	tempDir := t.TempDir()

	// Setup for permission error test (read-only directory)
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0500); err != nil {
		t.Fatal(err)
	}

	// Setup for PathExists error (file where dir should be)
	fileAsDir := filepath.Join(tempDir, "file")
	if err := os.WriteFile(fileAsDir, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		dirPath       string
		expectCreated bool
		expectError   bool
	}{
		{
			name:          "Directory already exists",
			dirPath:       tempDir,
			expectCreated: false,
			expectError:   false,
		},
		{
			name:          "Directory does not exist",
			dirPath:       filepath.Join(tempDir, "new-subdir"),
			expectCreated: true,
			expectError:   false,
		},
		{
			name:          "Permission denied (Mkdir fails)",
			dirPath:       filepath.Join(readOnlyDir, "fail-subdir"),
			expectCreated: false,
			expectError:   true,
		},
		{
			name:          "Path error (PathExists fails)",
			dirPath:       filepath.Join(fileAsDir, "impossible-subdir"),
			expectCreated: false,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			created, err := CreateDirIfNotExists(tc.dirPath)

			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if created != tc.expectCreated {
				t.Errorf("Expected created=%v, got %v", tc.expectCreated, created)
			}
		})
	}
}
