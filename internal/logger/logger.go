package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type flusher interface {
	Flush() error
}

type Logger struct {
	*logrus.Logger
	outputDir string
	flusher   flusher
}

func NewLogger(verbose bool) *Logger {
	log := logrus.New()
	
	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	
	log.SetOutput(os.Stdout)
	
	return &Logger{
		Logger: log,
	}
}

func (l *Logger) SetOutputDirectory(outputDir string) {
	l.outputDir = outputDir
	
	// Create log directory
	logsDir := filepath.Join(outputDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		l.Errorf("Failed to create log directory: %v", err)
		return
	}
	
	// Create log file with better permissions
	logFile := filepath.Join(logsDir, "panoptic.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.Errorf("Failed to create log file: %v", err)
		return
	}
	
	// Wrap file in buffered writer for performance
	bufferedWriter := bufio.NewWriterSize(file, 64*1024) // 64KB buffer
	
	// Set up flush on program exit
	go func() {
		// This goroutine will ensure buffer is flushed periodically
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			if err := bufferedWriter.Flush(); err != nil {
				// Log to stderr since file might have issues
				l.Errorf("Failed to flush log buffer: %v", err)
			}
		}
	}()
	
	l.SetOutput(bufferedWriter)
	l.Infof("Log file: %s (buffered)", logFile)
	
	// Store flusher for testing
	l.flusher = bufferedWriter
}

// Flush flushes the log buffer if it exists
func (l *Logger) Flush() error {
	if l.flusher != nil {
		return l.flusher.Flush()
	}
	return nil
}