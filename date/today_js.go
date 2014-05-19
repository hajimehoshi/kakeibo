// +build js

package date

import (
	"time"
	"github.com/gopherjs/gopherjs/js"
)

func today() Date {
	// In JavaScript, time.Now() returns a time which locale is UTC and
	// doesn't include the locale info. Use JavaScript's Date class instead.
	jsToday := js.Global.Get("Date").New()
	year := jsToday.Call("getFullYear").Int()
	month := time.Month(jsToday.Call("getMonth").Int() + 1)
	day := jsToday.Call("getDate").Int()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return Date(today.Unix())
}
