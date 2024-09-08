// Logic for organising files
package org

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/RMBeristain/organise-downloads/internal/common"
	"github.com/RMBeristain/organise-downloads/internal/logging"
	"github.com/RMBeristain/organise-downloads/local_utils"
)

var contains = local_utils.Contains
var dieIf = common.DieIf

// GetFilesToMove return a map of subdirs to slices of files.
//
// - files is a slice of DirEntries that should be moved.
// - excludedExtensions is the address of a string slice containing file or dir names that must not be moved.
//
// Each targets key is a destination subdir, and its value is a slice of the files that should be moved into it.
func GetFilesToMove(files []fs.DirEntry, excludedExtensions *[]string) (targets map[string][]string) {
	targets = make(map[string][]string)
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			if _, ok := targets[fileName]; !ok {
				targets[fileName] = []string{}
			}
		} else {
			fileExtension, destination := common.GetExtAndSubdir(fileName)

			if contains(*excludedExtensions, fileExtension) {
				continue
			}
			targets[destination] = append(targets[destination], fileName)
		}
	}
	return targets
}

// MoveFiles sequentially moves each file to its corresponding directory.
func MoveFiles(sourcePath string, filesToMove map[string][]string, logger *logging.Logr) {
	var movedFileCount int = 0
	var totalFileCount int = 0
	logr := *logger

	for subDir, files := range filesToMove {
		batchSize := len(files)
		totalFileCount += batchSize
		if batchSize > 0 {
			logr.LogInfo.Printf("Working on %v %v", batchSize, subDir)
			for i, file := range files {
				srcFilePath := filepath.Join(sourcePath, file)
				dstSubDir := filepath.Join(sourcePath, subDir)
				dstFilePath := filepath.Join(dstSubDir, file)

				logr.LogInfo.Printf("...moving %02d %v -> %v", i+1, srcFilePath, dstFilePath)

				if exists, err := common.PathExists(dstFilePath); !exists && err == nil {
					_, err := common.CreateDirIfNotExists(dstSubDir)
					dieIf(err)
					// TODO: 2024-09-08 acquire file lock to prevent race conditions
					dieIf(os.Rename(srcFilePath, dstFilePath))
				} else if exists {
					logr.LogError.Printf("Skipping file '%v' that already exists in: %v", file, dstFilePath)
				} else {
					logr.LogFatal.Print(err)
				}
				movedFileCount += 1
			}
		}
	}
	logr.LogInfo.Printf("Moved %v/%v files into %v subdirs.\n", movedFileCount, totalFileCount, len(filesToMove))
}
