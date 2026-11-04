package session

import "time"

type Session string

const (
	Pre     Session = "pre"
	Regular Session = "regular"
	Post    Session = "post"
	Closed  Session = "closed"
)

var ET = mustLoad()

func mustLoad() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	return loc
}

// At classifies t into pre (4:00-9:30), regular (9:30-16:00), post (16:00-20:00) ET.
func At(t time.Time) Session {
	t = t.In(ET)

	if wd := t.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return Closed
	}

	m := t.Hour()*60 + t.Minute()

	switch {
	case m >= 4*60 && m < 9*60+30:
		return Pre
	case m >= 9*60+30 && m < 16*60:
		return Regular
	case m >= 16*60 && m < 20*60:
		return Post
	}

	return Closed
}

// PreOpen returns 4:00 ET on t's date.
func PreOpen(t time.Time) time.Time {
	y, m, d := t.In(ET).Date()

	return time.Date(y, m, d, 4, 0, 0, 0, ET)
}

// LastClosedDay returns the latest weekday whose after-hours session has ended by t.
func LastClosedDay(t time.Time) time.Time {
	t = t.In(ET)

	y, m, d := t.Date()
	day := time.Date(y, m, d, 0, 0, 0, 0, ET)

	if t.Hour() < 20 {
		day = day.AddDate(0, 0, -1)
	}

	for day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		day = day.AddDate(0, 0, -1)
	}

	return day
}

func Date(t time.Time) string { return t.In(ET).Format("2006-01-02") }
