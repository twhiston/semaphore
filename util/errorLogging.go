package util

import (
	log "github.com/Sirupsen/logrus"
)

// LogWarning logs a warning with arbitrary field if error
func LogWarning(err error) bool {
	return LogWarningWithFields(err, log.Fields{"level": "Warn"})
}

// LogWarningWithFields logs a warning with added field context if error
func LogWarningWithFields(err error, fields log.Fields) bool {
	if err != nil {
		log.WithFields(fields).Warn(err.Error())
		return true
	}
	return false
}

// LogError logs an error with arbitrary field if error
func LogError(err error) bool {
	return LogErrorWithFields(err, log.Fields{"level": "Error"})
}

// LogErrorWithFields logs a error with added field context if error
func LogErrorWithFields(err error, fields log.Fields) bool {
	if err != nil {
		log.WithFields(fields).Error(err.Error())
		return true
	}
	return false
}

// LogPanic logs and panics with arbitrary field if error
func LogPanic(err error) bool {
	return LogPanicWithFields(err, log.Fields{"level": "Panic"})
}

// LogPanicWithFields logs and panics with added field context if error
func LogPanicWithFields(err error, fields log.Fields) bool {
	if err != nil {
		log.WithFields(fields).Panic(err.Error())
		return true
	}
	return false
}
