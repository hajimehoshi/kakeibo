// +build js

package date

import (
	"time"
	"github.com/gopherjs/gopherjs/js"
)

func Today() Date {
	// time.Now() doesn't work well because this needs access to the zone file.
	// Use JavaScript's Date class instead.
	jsToday := js.Global.Get("Date").New()
	year := jsToday.Call("getFullYear").Int()
	month := time.Month(jsToday.Call("getMonth").Int() + 1)
	day := jsToday.Call("getDate").Int()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return Date{today}
}
