package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	outputDir string
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
	
	// Create log file
	logFile := filepath.Join(logsDir, "panoptic.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.Errorf("Failed to create log file: %v", err)
		return
	}
	
	l.SetOutput(file)
	l.Infof("Log file: %s", logFile)
}