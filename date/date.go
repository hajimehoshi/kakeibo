package date

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"time"
)

type Date struct {
	t time.Time
}

func New(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func ParseISO8601(value string) (Date, error) {
	const layout = "2006-01-02"
	t, err := time.Parse(layout, value)
	if err != nil {
		return Date{}, err
	}
	return Date{t}, err
}

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

func (d Date) AddDate(years, months, days int) Date {
	return Date{d.t.AddDate(years, months, days)}
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.t.Year(), d.t.Month(), d.t.Day())
}
