// +build js

package date

import (
	"github.com/gopherjs/gopherjs/js"
	"time"
)

func today() Date {
	// In JavaScript, time.Now() returns a time whose locale is UTC and
	// doesn't include the locale info. Use JavaScript's Date class instead.
	jsToday := js.Global.Get("Date").New()
	year := jsToday.Call("getFullYear").Int()
	month := time.Month(jsToday.Call("getMonth").Int() + 1)
	day := jsToday.Call("getDate").Int()
	return New(year, month, day)
}
