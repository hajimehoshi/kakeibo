package date

import (
	"fmt"
	"time"
)

type Date struct {
	t time.Time
}

func New(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
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
	return Date{t}, err
}

func (d Date) AddDate(years, months, days int) Date {
	return Date{d.t.AddDate(years, months, days)}
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.t.Year(), d.t.Month(), d.t.Day())
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
