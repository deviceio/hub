package logging

import (
	"fmt"
	"log"
	"runtime/debug"
)

// DefaultLogger simply logs to the standard golang logger
type DefaultLogger struct{}

// Info ...
func (t *DefaultLogger) Info(format string, args ...interface{}) {
	log.Println("INFO: " + fmt.Sprintf(format, args...))
}

// Warn ...
func (t *DefaultLogger) Warn(format string, args ...interface{}) {
	log.Println("WARN: " + fmt.Sprintf(format, args...))
}

// Error ...
func (t *DefaultLogger) Error(format string, args ...interface{}) {
	log.Println("ERROR: "+fmt.Sprintf(format, args...), string(debug.Stack()))
}

// Debug ...
func (t *DefaultLogger) Debug(format string, args ...interface{}) {
	log.Println("DEBUG: " + fmt.Sprintf(format, args...))
}

// Fatal ...
func (t *DefaultLogger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(
		"%v %v \r\n %v",
		"FATAL:",
		fmt.Sprintf(format, args...),
		debug.Stack(),
	)
	log.Fatal(msg)
}
