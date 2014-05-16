package date_test

import (
	. "github.com/hajimehoshi/kakeibo/date"
	"testing"
)

func TestZero(t *testing.T) {
	expected := New(1970, 1, 1)
	d := Date{}
	if expected != d {
		t.Errorf("expected %+v got %+v", expected, d)
	}
}
