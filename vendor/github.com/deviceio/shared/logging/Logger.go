package logging

// Logger provides a generic logger interface across the whole application
type Logger interface {
	// Info logs a message with INFO priority
	Info(string, ...interface{})

	// Warn logs a message with WARN priority
	Warn(string, ...interface{})

	// Error logs a message with ERROR priority
	Error(string, ...interface{})

	// Debug logs a message with DEBUG priority
	Debug(string, ...interface{})

	// Fatal logs a message with FATAL priority and panics
	Fatal(string, ...interface{})
}
