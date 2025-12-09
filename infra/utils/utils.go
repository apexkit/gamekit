package utils

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func IsNoRecord(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func TrackTime(key string, pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	fmt.Println(key, elapsed)
	return elapsed
}

// 根据时区获取时间
func GetCurrentTimeByTimezone(timezone string) time.Time {
	if timezone == "" {
		return time.Now()
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Now()
	}

	return time.Now().In(location)
}

func GetTimeTimezone(t time.Time, timezone string) time.Time {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return t
	}

	return t.In(location)
}
