package date_test

import (
	. "github.com/hajimehoshi/kakeibo/date"
	"testing"
)

func TestZero(t *testing.T) {
	expected := New(1, 1, 1)
	d := Date(0)
	if expected != d {
		t.Errorf("expected %+v got %+v", expected, d)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		Date     Date
		Expected string
	}{
		{
			New(1, 1, 1),
			"0001-01-01",
		},
		{
			New(1234, 5, 6),
			"1234-05-06",
		},
		{
			New(1600, 10, 21),
			"1600-10-21",
		},
		{
			New(1969, 12, 31),
			"1969-12-31",
		},
		{
			New(1970, 1, 1),
			"1970-01-01",
		},
		{
			New(2006, 1, 5),
			"2006-01-05",
		},
		{
			New(9999, 12, 31),
			"9999-12-31",
		},
	}

	for _, test := range tests {
		if test.Expected != test.Date.String() {
			t.Errorf(
				"expected %+v got %+v",
				test.Expected,
				test.Date.String())
		}
	}
}
