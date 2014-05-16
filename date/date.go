package date

import (
	"fmt"
	"time"
)

type Date struct {
	unixTime int64
}

func New(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()}
}

func Today() Date {
	return today()
}

func ParseISO8601(value string) (Date, error) {
	const layout = "2006-01-02"
	t, err := time.Parse(layout, value)
	if err != nil {
		return Date{}, err
	}
	return Date{t.Unix()}, err
}

func (d Date) AddDate(years, months, days int) Date {
	t := time.Unix(d.unixTime, 0)
	return Date{t.AddDate(years, months, days).Unix()}
}

func (d Date) String() string {
	t := time.Unix(d.unixTime, 0)
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(text []byte) (err error) {
	*d, err = ParseISO8601(string(text))
	return
}
