package db

import (
	"time"
	log "github.com/Sirupsen/logrus"
)

// DatabaseTimeFormat represents the format that dredd uses to validate the datetime.
// This is not the same as the raw value we pass to a new object so
// we need to use this to coerce raw values to meet the API standard
// /^\d{4}-(?:0[0-9]{1}|1[0-2]{1})-[0-9]{2}T\d{2}:\d{2}:\d{2}Z$/
const DatabaseTimeFormat = "2006-01-02T15:04:05:99Z"

// GetParsedTime returns the timestamp as it will retrieved from the database
// This allows us to create timestamp consistency on return values from create requests
func GetParsedTime(t time.Time) time.Time {
	parsedTime, err := time.Parse(DatabaseTimeFormat, t.Format(DatabaseTimeFormat))
	if err != nil {
		log.Error(err)
	}
	return parsedTime
}