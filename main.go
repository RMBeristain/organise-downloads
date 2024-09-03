package main

import (
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/RMBeristain/organise-downloads/local_utils"
)

var (
	defaultSrcDir string = "Downloads"
	logFileName   string = "organise-downloads.log"
	logDirName    string = "log_files"
	LogDebug      *log.Logger
	LogInfo       *log.Logger
	LogError      *log.Logger
	LogFatal      *log.Logger
	logLevel      int = LogLevelInfo
	Contains          = local_utils.Contains
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func initLoggingToFile() {
	var logDir string
	var logFile string

	pDownloadDir := flag.String("downloads", defaultSrcDir, "Full path to Downloads dir")
	pNewLogLevel := flag.Int("loglevel", LogLevelInfo, "Use this log level [0:3]")
	flag.Parse() // read command line flags

	if *pNewLogLevel != logLevel && LogLevelDebug <= *pNewLogLevel && *pNewLogLevel <= LogLevelError {
		logLevel = *pNewLogLevel
	}

	if *pDownloadDir != defaultSrcDir {
		defaultSrcDir = *pDownloadDir // use command line value
	} else {
		defaultSrcDir = *getCurrentUserDownloadPath()
	}

	log.Printf("Will log to %v\n", defaultSrcDir)
	logDir = filepath.Join(defaultSrcDir, logDirName)

	if created, err := createDirIfNotExists(logDir); err != nil {
		log.Fatalf("FATAL: Unable to create dir %v - %v", logDir, err)
	} else if created {
		log.Printf("Created missing directory %v", logDir)
	}

	logFile = filepath.Join(logDir, logFileName)

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panicf("Cannot write to log file %v - %v", logFile, err)
	}

	LogDebug = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Llongfile)
	LogInfo = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogError = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	LogFatal = log.New(file, "FATAL: ", log.Ldate|log.Ltime|log.Llongfile)
}

func main() {
	initLoggingToFile()

	LogInfo.Print("START.")
	files, err := os.ReadDir(defaultSrcDir) // get all files
	dieIf(err)

	filesToMove := getFilesToMove(files)

	if len(filesToMove) > 0 {
		LogInfo.Printf("Files to move : %v\n", filesToMove)
		moveFiles(defaultSrcDir, filesToMove)
	} else {
		LogInfo.Print("No files to move.")
	}

	LogInfo.Print("DONE.")
	os.Exit(0)
}

// Returns true if dir was created, else false; if there is an error returns (false, err)
func createDirIfNotExists(dirName string) (bool, error) {
	if exists, err := pathExists(dirName); !exists && err == nil {
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

// Check whether there was an error.
//
// If an error exists, terminate.
func dieIf(err error) {
	if err != nil {
		LogFatal.Printf("something went wrong: %v", err.Error())
		log.Fatalf("FATAL: %v", err) // panic(err)
	}
}

// Returns whether the given file or directory exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		log_debug("Path %v doesn't exist.", path)
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

// Return a map of subdirs to slices of files.
//
// Each key is a destination subdir, and its value is a slice of files that should be moved into it.
func getFilesToMove(files []fs.DirEntry) (targets map[string][]string) {
	targets = make(map[string][]string)
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() {
			if _, ok := targets[fileName]; !ok {
				targets[fileName] = []string{}
			}
		} else {
			fileExtension, destination := getExtAndSubdir(fileName)

			if Contains(excludedExtensions(), fileExtension) {
				continue
			}
			targets[destination] = append(targets[destination], fileName)
		}
	}
	return targets
}

// Create new subdirs if they not exist. If a subdir cannot be created, panic.
//
// - subDir: the new subdir to be created
func createSubdirIfNotExists(subDir string) {
	if exists, err := pathExists(subDir); err == nil && !exists {
		err := os.Mkdir(subDir, 0777)
		if err != nil {
			log.Printf("Is 'subDir'='%v' a fully qualified path?", subDir)
			log.Fatalf("unable to create dir %v: %v", subDir, err.Error())
		}
		log.Printf("created dir %v", subDir)
	} else if err != nil {
		log.Fatalf(err.Error())
	}
}

// Move each file to its corresponding directory.
func moveFiles(sourcePath string, filesToMove map[string][]string) {
	var movedFileCount int = 0
	var totalFileCount int = 0

	for newPath, files := range filesToMove {
		totalFileCount += len(files)
		for _, file := range files {
			oldFullPath := filepath.Join(sourcePath, file)
			subDir := filepath.Join(sourcePath, newPath)
			newFullPath := filepath.Join(subDir, file)

			LogInfo.Printf("...moving %v -> %v", oldFullPath, newFullPath)

			if exists, err := pathExists(newFullPath); !exists && err == nil {
				createSubdirIfNotExists(subDir)
				dieIf(os.Rename(oldFullPath, newFullPath))
			} else if exists {
				LogError.Printf("Skipping file '%v' that already exists in: %v", file, newFullPath)
			} else {
				LogFatal.Print(err)
			}

			movedFileCount += 1
		}

	}
	LogInfo.Printf("Moved %v/%v files into %v subdirs.\n", movedFileCount, totalFileCount, len(filesToMove))
}

// Log DEBUG level message if logLevel is set to `LogLevelDebug`
func log_debug(message string, args ...interface{}) {
	if logLevel == LogLevelDebug {
		defer func() {
			if r := recover(); r != nil {
				// LogDebug isn't defined yet
				log.Printf(message, args...)
			}
		}()
		LogDebug.Printf(message, args...)
	}
}

// getCurrentUserDownloadPath finds the current user and their home directory. The return value is the address of a
// string variable that stores the value of the fully-qualified path to 'Downloads' dir (e.g. /Users/me/Downloads).
func getCurrentUserDownloadPath() *string {
	currentUser, err := user.Current()
	dieIf(err)
	homeDir, err := os.UserHomeDir()
	dieIf(err)
	defaultPath := filepath.Join(homeDir, defaultSrcDir)
	log_debug("Using default path %v for user %v\n", defaultPath, currentUser.Username) // use log; LogInfo isn't ready
	return &defaultPath
}
