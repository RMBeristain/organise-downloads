// Shared logic
package common

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// GetCurrentUserDownloadPath finds the current user and their home directory. The return value is the address of a
// string variable that stores the value of the fully-qualified path to 'Downloads' dir (e.g. /Users/me/Downloads).
func GetCurrentUserDownloadPath(defaultSrcDir string) *string {
	currentUser, err := user.Current()
	if err != nil {
		log.Panicf("Unable to determine current user: %v", err.Error())
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Unable to determine %v home dir: %v", currentUser, err.Error())
	}

	defaultPath := filepath.Join(homeDir, defaultSrcDir)
	// log_debug("Using default path %v for user %v\n", defaultPath, currentUser.Username) // use log; LogInfo isn't ready
	return &defaultPath
}

// PathExists returns whether the given file or directory exists
func PathExists(path string) (exists bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// CreateDirIfNotExists returns true if dir was created, else false; if there is an error returns (false, err)
func CreateDirIfNotExists(dirName string) (wasCreated bool, err error) {
	if exists, err := PathExists(dirName); !exists && err == nil {
		log.Printf("Dir %v does not exist, will try to create it...\n", dirName)
		err = os.Mkdir(dirName, 0777)
		if err != nil {
			log.Printf("Unable to create dir %v - %v\n", dirName, err)
			return false, err
		}

		return true, nil
	} else if err != nil {
		log.Printf("Unable to create log directory %v", err)
		return false, err
	}

	return false, nil
}

// DieIf checks whether there was an error. If an error exists, log it and terminate.
func DieIf(err error) {
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

// GetExtAndSubdir returns a file's extension and corresponding target subdir.
func GetExtAndSubdir(fileName string) (fileExtension, subDirName string) {
	fileExtension = filepath.Ext(fileName)
	return fileExtension, strings.Replace(fileExtension, ".", "", 1) + "_files"
}
