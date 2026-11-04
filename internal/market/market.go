package market

import "time"

// Bar is one minute aggregate; Start and End are unix milliseconds.
type Bar struct {
	Sym   string
	C     float64
	V     float64
	OP    float64
	Start int64
	End   int64
}

func (b Bar) StartTime() time.Time { return time.UnixMilli(b.Start).UTC() }

func (b Bar) EndTime() time.Time { return time.UnixMilli(b.End).UTC() }
