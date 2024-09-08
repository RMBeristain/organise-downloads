// Logging package
package logging

import (
	"log"
	"os"
	"path/filepath"

	"github.com/RMBeristain/organise-downloads/internal/common"
	"github.com/RMBeristain/organise-downloads/local_utils"
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var (
	WorkingSrcDir string = "Downloads"
	LogFileName   string = "organise-downloads.log"
	LogDirName    string = "log_files"
	LogLevel      int    = LogLevelInfo
	Contains             = local_utils.Contains
)

type Logr struct {
	logDebug *log.Logger
	LogInfo  *log.Logger
	LogError *log.Logger
	LogFatal *log.Logger
}

// log_debug logs a DEBUG level message if LogLevel is set to `LogLevelDebug`
func LogDebug(logr *Logr, message string, args ...interface{}) {
	if LogLevel == LogLevelDebug {
		defer func() {
			if r := recover(); r != nil {
				// LogDebug isn't defined yet
				log.Printf(message, args...)
			}
		}()
		logr.logDebug.Printf(message, args...)
	}
}

// InitLoggingToFile creates or prepares a file for logging and returns the address of a configured Logr struct.
func InitLoggingToFile(WorkingSrcDir *string, pNewLogLevel *int) *Logr {
	var logDir string
	var logFile string
	var logr Logr

	if *pNewLogLevel != LogLevel && LogLevelDebug <= *pNewLogLevel && *pNewLogLevel <= LogLevelError {
		LogLevel = *pNewLogLevel
	}

	logDir = filepath.Join(*WorkingSrcDir, LogDirName)

	if created, err := common.CreateDirIfNotExists(logDir); err != nil {
		log.Fatalf("FATAL: Unable to create dir %v - %v", logDir, err)
	} else if created {
		log.Printf("Created missing directory %v", logDir)
	}

	logFile = filepath.Join(logDir, LogFileName)

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panicf("Cannot write to log file %v - %v", logFile, err)
	}

	logr.logDebug = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Llongfile)
	logr.LogInfo = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logr.LogError = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logr.LogFatal = log.New(file, "FATAL: ", log.Ldate|log.Ltime|log.Llongfile)

	LogDebug(&logr, "Writing log to %v", logDir)

	return &logr
}
