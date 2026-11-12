// helper-------------------------------------
// @file      : time.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/25 15:04
// -------------------------------------------

package helper

import (
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"time"
)

// GetNowTimeString returns the current time string for the given timezone
func GetNowTimeString() string {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {
		// If the timezone fails to load, fall back to UTC as the default
		loc = time.UTC
	}
	now := time.Now().In(loc)
	return now.Format("2006-01-02 15:04:05")
}

func GetNowTime() time.Time {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {
		// If the timezone fails to load, fall back to UTC as the default
		loc = time.UTC
	}
	return time.Now().In(loc)
}

// FormatTime formats a time.Time into a string using the globally configured timezone
func FormatTime(t time.Time) string {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format("2006-01-02 15:04:05")
}
