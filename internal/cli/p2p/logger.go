package p2p

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/flootic/envseal/internal/cli/config"
)

var initLoggerOnce sync.Once

// newLogger creates a logger that writes to a file in the system's temp directory.
// The logger is lazily initialized on the first call to avoid unnecessary file creation.
func newLogger() *log.Logger {
	var logger *log.Logger
	initLoggerOnce.Do(func() {
		logFilePath := filepath.Join(os.TempDir(), config.Directory, "p2p.log")
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// If we can't create the log file, fallback to discarding logs instead of crashing.
			logger = log.New(io.Discard, "p2p: ", log.LstdFlags)
			return
		}
		logger = log.New(logFile, "p2p: ", log.LstdFlags)
	})
	return logger
}
