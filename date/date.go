package date

import (
	"fmt"
	"time"
)

const durationDay = 24 * time.Hour
var (
	unixEpoch = time.Unix(0, 0)
)

type Date int32

func New(year int, month time.Month, day int) Date {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	v := int64(t.Sub(time.Time{})) / int64(durationDay)
	return Date(v)
}

func Today() Date {
	return today()
}

func ParseISO8601(value string) (Date, error) {
	const layout = "2006-01-02"
	t, err := time.Parse(layout, value)
	if err != nil {
		return Date(0), err
	}
	return New(t.Year(), t.Month(), t.Day()), nil
}

func (d Date) time() time.Time {
	dur := time.Duration(int64(d) * int64(durationDay))
	fmt.Printf("%v, %v, %v\n", int32(d), int64(dur), time.Time{}.Add(dur).UTC())
	return time.Time{}.Add(dur).UTC()
}

func (d Date) AddDate(years, months, days int) Date {
	return Date(d.time().AddDate(years, months, days).Unix())
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year(), d.Month(), d.Day())
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(text []byte) (err error) {
	*d, err = ParseISO8601(string(text))
	return
}

func (d Date) Date() (year int, month time.Month, day int) {
	return d.Year(), d.Month(), d.Day()
}

func (d Date) Year() int {
	return d.time().Year()
}

func (d Date) Month() time.Month {
	return d.time().Month()
}

func (d Date) Day() int {
	return d.time().Day()
}
