// +build !js

package date

import (
	"time"
)

func today() Date {
	now := time.Now()
	today := time.Date(
		now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return Date(today.Unix())
}
