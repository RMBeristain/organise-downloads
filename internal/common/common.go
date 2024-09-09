// Shared logic
package common

import (
	"errors"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/RMBeristain/organise-downloads/internal/logging"
)

var logger = &logging.ConfiguredZerologger

// GetCurrentUserDownloadPath finds the current user and their home directory. The return value is the address of a
// string variable that stores the value of the fully-qualified path to 'Downloads' dir (e.g. /Users/me/Downloads).
func GetCurrentUserDownloadPath(defaultSrcDir string) *string {
	currentUser, err := user.Current()
	if err != nil {
		logger.Panic().Err(err).Msg("unable to determine current user")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Panic().Err(err).Str("currentUser", currentUser.Username).Msg("unable to determine home dir")
	}

	workingPath := filepath.Join(homeDir, defaultSrcDir)
	logger.Debug().Str("workingPath", workingPath).Str("currentUser", currentUser.Username).Send()
	return &workingPath
}

// PathExists returns whether the given file or directory exists
func PathExists(path string) (exists bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		logger.Trace().Str("path", path).Msg("already exists")
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		logger.Trace().Str("path", path).Err(err).Msg("doesn't exist")
		return false, nil
	}
	logger.Err(err).Msg("unexpected error")
	return false, err
}

// CreateDirIfNotExists returns true if dir was created, else false; if there is an error returns (false, err)
func CreateDirIfNotExists(dirName string) (wasCreated bool, err error) {
	if exists, err := PathExists(dirName); !exists && err == nil {
		logger.Debug().Str("targetDir", dirName).Msg("attempting to create missing dir")
		err = os.Mkdir(dirName, 0777)
		if err != nil {
			logger.Err(err).Str("targetDir", dirName).Msg("unable to create dir")
			return false, err
		}

		return true, nil
	} else if err != nil {
		logger.Err(err).Msg("unable to create log dir")
		return false, err
	}

	return false, nil
}

// DieIf checks whether there was an error. If an error exists, log it and terminate.
func DieIf(err error) {
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
}

// GetExtAndSubdir returns a file's extension and corresponding target subdir.
func GetExtAndSubdir(fileName string) (fileExtension, subDirName string) {
	fileExtension = filepath.Ext(fileName)
	return fileExtension, strings.Replace(fileExtension, ".", "", 1) + "_files"
}
