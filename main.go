package main

import (
	"flag"
	"os"

	"github.com/RMBeristain/organise-downloads/internal/common"
	"github.com/RMBeristain/organise-downloads/internal/logging"
	"github.com/RMBeristain/organise-downloads/internal/org"
)

var (
	defaultSrcDir string = "Downloads"
)

func main() {
	var pWorkingSrcDir *string

	pDownloadDir := flag.String("downloads", defaultSrcDir, "Full path to Downloads dir")
	pNewLogLevel := flag.Int("loglevel", logging.LogLevelInfo, "Use this log level [0:3]")
	flag.Parse() // read command line flags

	if *pDownloadDir != defaultSrcDir {
		pWorkingSrcDir = pDownloadDir // use command line value
	} else {
		pWorkingSrcDir = common.GetCurrentUserDownloadPath(defaultSrcDir)
	}

	logr := *logging.InitLoggingToFile(pWorkingSrcDir, pNewLogLevel)

	logr.LogInfo.Print("START.")
	files, err := os.ReadDir(*pWorkingSrcDir) // get all files
	if err != nil {
		logr.LogFatal.Fatalf("FATAL: %v", err)
	}

	filesToMove := org.GetFilesToMove(files, excludedExtensions())

	if len(filesToMove) > 0 {
		logr.LogInfo.Printf("Files to move : %v\n", filesToMove)
		org.MoveFiles(*pWorkingSrcDir, filesToMove, &logr)
	} else {
		logr.LogInfo.Print("No files to move.")
	}

	logr.LogInfo.Print("DONE.")
	os.Exit(0)
}

// Return a slice of extensions that won't be moved into subdirs.
func excludedExtensions() *[]string {
	return &[]string{".DS_Store", ".localized"}
}
