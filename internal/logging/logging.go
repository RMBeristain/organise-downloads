// Logging package
package logging

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

var (
	defaultLogParentDir  string = "Downloads"
	LogFileName          string = "organise-downloads.log"
	LogDir               string = "log_files"
	ConfiguredZerologger zerolog.Logger
)

type Zerologger struct {
	zerolog.Logger
}

// InitZeroLog configures a zerolog Logger and returns an instance of the Zerologr struct that contains it.
func InitZeroLog() Zerologger {
	userHomeDir, _ := os.UserHomeDir()
	logDirPath := filepath.Join(userHomeDir, defaultLogParentDir, LogDir)

	// check if logDirPath exists; create it if not.
	_, err := os.Stat(logDirPath)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.Mkdir(logDirPath, 0777)
		if err != nil {
			log.Fatalf("unable to create logging dir %v", logDirPath)
		}
	}

	logFilePath := filepath.Join(logDirPath, LogFileName)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panicf("Cannot write to log file %v - %v", logFilePath, err)
	}

	ConfiguredZerologger = zerolog.New(file).With().Timestamp().Caller().Logger()
	return Zerologger{ConfiguredZerologger}
}
