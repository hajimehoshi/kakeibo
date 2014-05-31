// +build !js

package date

import (
	"time"
)

func today() Date {
	now := time.Now()
	return New(now.Year(), now.Month(), now.Day())
}
