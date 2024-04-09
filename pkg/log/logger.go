package chatsnippet

import (
	"fmt"
	"log"
	"os"
	cfg "protocolized_chat/pkg/config"
)

type CustomLogger struct {
	logger  *log.Logger
	logFile *os.File
}

func (cl *CustomLogger) Close() error {
	if err := cl.logFile.Close(); err != nil {
		return fmt.Errorf("closing log file: %v", err)
	}

	return nil
}

func GetLogger() *CustomLogger {
	file, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("creating log file: %v", err)
	}

	logger := log.New(file, "", log.LUTC|log.Ldate|log.Lshortfile|log.Ltime)

	return &CustomLogger{logger: logger, logFile: file}
}

func (cl *CustomLogger) Println(args ...any) {
	cl.logger.Println(args...)
}
