package common

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestGenerateSampleToml(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		validate  func(t *testing.T, path string)
		expectErr bool
	}{
		{
			name: "Happy Path - Create valid TOML file (explicit filename)",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "default_config.toml")
			},
			validate: func(t *testing.T, path string) {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Fatalf("File was not created at %s", path)
				}
				got, err := LoadExcludedExtensions(path)
				if err != nil {
					t.Fatalf("Failed to load generated file: %v", err)
				}
				if !reflect.DeepEqual(got, DefaultExcludedExtensions) {
					t.Errorf("Expected %v, got %v", DefaultExcludedExtensions, got)
				}
			},
			expectErr: false,
		},
		{
			name: "Happy Path - Create valid TOML file (directory only)",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, path string) {
				expectedPath := filepath.Join(path, "sampleOrganiseDownloads.toml")
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Fatalf("File was not created at %s", expectedPath)
				}
				got, err := LoadExcludedExtensions(expectedPath)
				if err != nil {
					t.Fatalf("Failed to load generated file: %v", err)
				}
				if !reflect.DeepEqual(got, DefaultExcludedExtensions) {
					t.Errorf("Expected %v, got %v", DefaultExcludedExtensions, got)
				}
			},
			expectErr: false,
		},
		{
			name: "Error Path - File already exists (directory provided)",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				filePath := filepath.Join(dir, "sampleOrganiseDownloads.toml")
				if err := os.WriteFile(filePath, []byte("existing content"), 0644); err != nil {
					t.Fatalf("unable to create existing file: %v", err)
				}
				return dir
			},
			expectErr: true,
		},
		{
			name: "Error Path - PathExists fails (directory no execute permission)",
			setup: func(t *testing.T) string {
				// Create a directory with read/write but NO execute permissions (0600)
				// This allows os.Stat(dir) to succeed, but os.Stat(dir/file) to fail with EACCES
				dir := filepath.Join(t.TempDir(), "no_exec_dir")
				if err := os.Mkdir(dir, 0600); err != nil {
					t.Fatalf("unable to create dir: %v", err)
				}
				// Restore permissions after test so cleanup works
				t.Cleanup(func() { os.Chmod(dir, 0755) })
				return dir
			},
			expectErr: true,
		},
		{
			name: "Error Path - Cannot create file (directory missing)",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing", "config.toml")
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup(t)
			err := GenerateSampleToml(path)

			if tc.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tc.validate != nil {
					tc.validate(t, path)
				}
			}
		})
	}
}

func TestLoadExcludedExtensions(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string // Returns the config path to use
		expected  []string
		expectErr bool
	}{
		{
			name: "Happy Path - No flag passed (use defaults)",
			setup: func(t *testing.T) string {
				return ""
			},
			expected:  DefaultExcludedExtensions, // Expect current hardcoded defaults
			expectErr: false,
		},
		{
			name: "Happy Path - Flag passed with valid TOML",
			setup: func(t *testing.T) string {
				// TOML with header 'excludedFiles' and newline separated values
				content := `excludedFiles = [
".mp3",
".mp4"
]`
				path := filepath.Join(t.TempDir(), "config.toml")
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("unable to write test file: %v", err)
				}
				return path
			},
			expected:  []string{".mp3", ".mp4"},
			expectErr: false,
		},
		{
			name: "Error Path - File does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "non_existent.toml")
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "Error Path - Invalid TOML content",
			setup: func(t *testing.T) string {
				content := `INVALID TOML CONTENT`
				path := filepath.Join(t.TempDir(), "bad_config.toml")
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("unable to write test file: %v", err)
				}
				return path
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configPath := tc.setup(t)
			got, err := LoadExcludedExtensions(configPath)

			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
