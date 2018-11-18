package migration

import "log"

var logger *log.Logger

// SetLogger will set the logger to be used during migrations
func SetLogger(l *log.Logger) {
	logger = l
}

func logPrintf(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf(format, args...)
	}
}
