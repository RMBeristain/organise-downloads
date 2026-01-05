package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/RMBeristain/organise-downloads/internal/common"
	"github.com/RMBeristain/organise-downloads/internal/logging"
	"github.com/RMBeristain/organise-downloads/internal/org"
	"github.com/rs/zerolog"
)

var (
	defaultSrcDir string = "Downloads"
)

func main() {
	var workingSrcDir string

	startTime := time.Now()

	pDownloadDir := flag.String("downloads", defaultSrcDir, "Full path to Downloads dir")
	pNewLogLevel := flag.Int("loglevel", int(zerolog.InfoLevel), "Use this log level [0:3]")
	pExcludedExtensions := flag.String("excludeExtensions", "", "Path to TOML file with excluded extensions")
	pGenerateSample := flag.String("generateSampleTomlFile", "", "Generate a sample TOML file at the specified path and exit")
	flag.Parse() // read command line flags

	if int(zerolog.TraceLevel) <= *pNewLogLevel && *pNewLogLevel <= int(zerolog.PanicLevel) {
		zerolog.SetGlobalLevel(zerolog.Level(*pNewLogLevel))
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	logger := logging.InitZeroLog()
	logger.Trace().Int("GlobalLogLevel", *pNewLogLevel).Msg("set new log level")

	if *pGenerateSample != "" {
		if err := common.GenerateSampleToml(*pGenerateSample); err != nil {
			logger.Fatal().Err(err).Msg("unable to generate sample TOML file")
		}
		logger.Info().Str("path", *pGenerateSample).Msg("generated sample TOML file")
		return
	}

	if *pDownloadDir != defaultSrcDir {
		logger.Debug().Str("downloadDir", *pDownloadDir).Msg("changed source dir")
		workingSrcDir = *pDownloadDir // use command line value
	} else {
		var err error
		workingSrcDir, err = common.GetCurrentUserDownloadPath(defaultSrcDir)
		if err != nil {
			logger.Fatal().Err(err).Msg("unable to determine downloads directory")
		}
	}

	logger.Info().Msg("START.")
	files, err := os.ReadDir(workingSrcDir) // get all files

	if err != nil {
		logger.Fatal().Err(err).Msg(err.Error())
	}

	filesChannel := make(chan string, 4)
	excluded, err := common.LoadExcludedExtensions(*pExcludedExtensions)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load excluded extensions")
	}
	filesToMove := org.GetFilesToMove(files, excluded)

	if len(filesToMove) > 0 {
		logger.Debug().Str("filesToMove", fmt.Sprintf("%v", filesToMove))
		go org.MoveFiles(workingSrcDir, filesToMove, filesChannel)
		for fileMoved := range filesChannel {
			logger.Info().Str("filePath", fileMoved).Msg("new location")
		}
	} else {
		logger.Info().Msg("No files to move.")
	}

	logger.Info().Dur("elapsedTime", time.Since(startTime)).Msg("DONE.")
}
