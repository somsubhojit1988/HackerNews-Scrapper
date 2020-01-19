package logger

import "log"

// Loglevel indicaes the severity of the log information
type Loglevel uint8

// Logger cntains an actual log.Logger and some flag to determine the logging severity
type Logger struct {
	logger *log.Logger
}

// logmsg is a container for each log msg string, and the log severity level
type logMsg struct {
	tag   string
	msg   string
	level Loglevel
}

const (
	// Debug to flag debug level logs
	Debug Loglevel = 0

	// Info flags an info level log
	Info Loglevel = 1

	// Verbose flags a verbose level log
	Verbose Loglevel = 2

	// Warning flags a WARNING level log
	Warning Loglevel = 3

	// Error flags an error log
	Error Loglevel = 4

	// Fatal flags a fatal log, and panics after the msg is published
	Fatal Loglevel = 5
)
