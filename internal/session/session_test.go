package session

import (
	"testing"
	"time"
)

func TestAt(t *testing.T) {
	cases := []struct {
		h, m int
		day  int
		want Session
	}{
		{3, 59, 2, Closed},
		{4, 0, 2, Pre},
		{9, 29, 2, Pre},
		{9, 30, 2, Regular},
		{15, 59, 2, Regular},
		{16, 0, 2, Post},
		{19, 59, 2, Post},
		{20, 0, 2, Closed},
		{10, 0, 7, Closed},
	}

	for _, c := range cases {
		if got := At(time.Date(2026, 11, c.day, c.h, c.m, 0, 0, ET)); got != c.want {
			t.Errorf("%02d:%02d day %d: got %s want %s", c.h, c.m, c.day, got, c.want)
		}
	}
}

func TestLastClosedDay(t *testing.T) {
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"tue after close before 20:00", time.Date(2026, 11, 3, 18, 0, 0, 0, ET), "2026-11-02"},
		{"tue after 20:00", time.Date(2026, 11, 3, 20, 0, 0, 0, ET), "2026-11-03"},
		{"mon morning", time.Date(2026, 11, 2, 8, 0, 0, 0, ET), "2026-10-30"},
		{"saturday", time.Date(2026, 11, 7, 12, 0, 0, 0, ET), "2026-11-06"},
		{"sunday night", time.Date(2026, 11, 8, 21, 0, 0, 0, ET), "2026-11-06"},
	}

	for _, c := range cases {
		if got := LastClosedDay(c.at).Format("2006-01-02"); got != c.want {
			t.Errorf("%s: got %s want %s", c.name, got, c.want)
		}
	}
}
