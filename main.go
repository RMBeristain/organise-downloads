package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	target   string = "/Users/rberistain/Downloads/"
	logDir   string = filepath.Join(target, "log_files")
	logFile  string = filepath.Join(logDir, "organise-downloads.log")
	LogDebug *log.Logger
	LogInfo  *log.Logger
	LogError *log.Logger
	LogFatal *log.Logger
	logLevel int = LogLevelInfo
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func init() {
	if exists, err := pathExists(logDir); !exists && err == nil {
		check(os.Mkdir(logDir, 0777))
	} else if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	LogDebug = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Llongfile)
	LogInfo = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogError = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogFatal = log.New(file, "FATAL: ", log.Ldate|log.Ltime|log.Llongfile)
}

func main() {
	LogInfo.Print("START.")
	files, err := ioutil.ReadDir(target)
	check(err)

	filesToMove, targetDirs := getFilesToMove(files)

	if len(filesToMove) > 0 {
		LogInfo.Printf("Files to move : %v\n", filesToMove)
		LogInfo.Printf("Target dirs   : %v\n", targetDirs)
		moveFiles(filesToMove, targetDirs)
	} else {
		LogInfo.Print("No files to move.")
	}

	LogInfo.Print("DONE.")
	os.Exit(0)
}

// Check whether there was an error.
//
// If an error exists, terminate.
func check(err error) {
	if err != nil {
		LogFatal.Print(err)
		log.Fatal(err) // panic(err)
	}
}

// Check whether 'slice' contains 'element'.
func contains(slice []string, element string) bool {
	for _, this := range slice {
		if this == element {
			return true
		}
	}
	return false
}

// Returns whether the given file or directory exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		log_debug("Path doesn't exist.")
		return false, nil
	}
	return false, err
}

// Return a copy of 'slice' without 'element'.
//
// How is this not built-in???
func delSliceElement(slice []string, element string) []string {
	newSlice := make([]string, len(slice))
	copy(newSlice, slice)
	for i, this := range newSlice {
		if this == element {
			return append(newSlice[:i], newSlice[i+1:]...)
		}
	}
	return slice
}

// Return a slice of extensions that won't be moved into subdirs.
func excludedExtensions() []string {
	return []string{".DS_Store", ".localized"}
}

// Return file's extension and corresponding target subdir for 'fileName'
func getExtAndSubdir(fileName string) (string, string) {
	fileExtension := filepath.Ext(fileName)
	return fileExtension, strings.Replace(fileExtension, ".", "", 1) + "_files"
}

// Return a slice of filenames that will be moved into corresponding subdirs, and a slice of subdirs.
//
// If subdir doesn't exist, create it.
func getFilesToMove(files []fs.FileInfo) ([]string, []string) {
	var existingDirs []string
	var newSubdirs []string
	var filesToMove []string

	for _, file := range files {
		if file.IsDir() {
			existingDirs = append(existingDirs, file.Name())
			log_debug("Found dir: %v\n", file.Name())

			if contains(newSubdirs, file.Name()) {
				newSubdirs = delSliceElement(newSubdirs, file.Name())
				log_debug("Subdir %v already exists.\n", file.Name())
			}
		} else {
			fileExtension, subdirName := getExtAndSubdir(file.Name())

			if contains(excludedExtensions(), fileExtension) {
				continue
			}
			filesToMove = append(filesToMove, file.Name())

			if contains(existingDirs, subdirName) || contains(newSubdirs, subdirName) {
				continue
			}

			newSubdirs = append(newSubdirs, subdirName)
			log_debug("Need Subdir: %v\n", subdirName)
		}
	}
	log_debug("Existing dirs:\t%v\n", existingDirs)

	for _, dir := range newSubdirs {
		check(os.Mkdir(target+dir, 0777))
		existingDirs = append(existingDirs, dir)
	}

	log_debug("New dirs:\t%v\n", newSubdirs)
	return filesToMove, existingDirs
}

// Move each file in 'files' to its corresponding directory in 'targetDirs'
func moveFiles(files []string, targetDirs []string) {
	var movedFiles int = 0

	for _, file := range files {
		_, subDir := getExtAndSubdir(file)

		if !contains(targetDirs, subDir) {
			LogError.Printf("Skipping file '%v' without corresponding subdir '%v'", file, subDir)
			continue
		}

		oldPath := filepath.Join(target, file)
		newPath := filepath.Join(target, subDir, file)

		LogInfo.Printf("...moving %v -> %v", oldPath, newPath)

		if exists, err := pathExists(newPath); !exists && err == nil {
			check(os.Rename(oldPath, newPath))
		} else if exists {
			LogError.Printf("Skipping file '%v' that already exists in: %v", file, newPath)
		} else {
			LogFatal.Print(err)
		}

		movedFiles += 1
	}

	LogInfo.Printf("Moved %v/%v files into %v subdirs.\n", movedFiles, len(files), len(targetDirs))
}

// Log DEBUG level message if logLevel is set to `LogLevelDebug`
func log_debug(message string, args ...interface{}) {
	if logLevel == LogLevelDebug {
		LogDebug.Printf(message, args...)
		bla("tesing only")
	}
}

func bla(param string) string {
	fmt.Println(param)
	return param
}
