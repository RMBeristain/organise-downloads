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

var (
	contains = local_utils.Contains
	logger   = &logging.ConfiguredZerologger
)

// GetFilesToMove return a map of subdirs to slices of files.
//
// - files is a slice of DirEntries that should be moved.
// - excludedExtensions is a slice of strings containing file or dir names that must not be moved.
//
// Each targets key is a destination subdir, and its value is a slice of the files that should be moved into it.
func GetFilesToMove(files []fs.DirEntry, excludedExtensions []string) (targets map[string][]string) {
	targets = make(map[string][]string)
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			if _, ok := targets[fileName]; !ok {
				targets[fileName] = []string{}
				logger.Trace().Str("fileName", fileName).Msg("found dir to process")
			}
		} else {
			fileExtension, destination := common.GetExtAndSubdir(fileName)

			if contains(excludedExtensions, fileExtension) {
				continue
			}
			targets[destination] = append(targets[destination], fileName)
		}
	}
	return targets
}

// MoveFiles sequentially moves each file to its corresponding directory.
func MoveFiles(sourcePath string, filesToMove map[string][]string, fileChannel chan string) {
	defer close(fileChannel)
	var movedFileCount int = 0
	var totalFileCount int = 0

	for subDir, files := range filesToMove {
		batchSize := len(files)
		totalFileCount += batchSize

		for i, file := range files {
			srcFilePath := filepath.Join(sourcePath, file)
			dstSubDir := filepath.Join(sourcePath, subDir)
			dstFilePath := filepath.Join(dstSubDir, file)

			if i == 0 {
				logger.Info().Int("batchSize", batchSize).Str("subDir", subDir).Msg("processing")
			}

			if isFileInUse(srcFilePath) {
				logger.Debug().Str("file", file).Msg("skipping file: currently in use")
				continue
			}

			if exists, err := common.PathExists(dstFilePath); !exists && err == nil {
				_, err := common.CreateDirIfNotExists(dstSubDir)
				if err != nil {
					logger.Err(err).Str("subDir", subDir).Msg("skipping batch: unable to create dir")
					continue
				}
				if err := os.Rename(srcFilePath, dstFilePath); err != nil {
					logger.Err(err).Str("file", file).Msg("skipping file: unable to rename")
					continue
				}
			} else if exists {
				logger.Err(err).Str("fileName", file).Str("dstFilePath", dstFilePath).Msg("skipped")
			} else {
				logger.Fatal().Err(err).Send()
			}
			movedFileCount += 1
			logger.Debug().Int("count", i+1).Str("srcFilePath", srcFilePath).Str("dstFilePath", dstFilePath).Msg("moved")
			fileChannel <- dstFilePath
		}
	}
	logger.Info().Int("movedCount", movedFileCount).Int("totalCount", totalFileCount).Msg("moved")
}
