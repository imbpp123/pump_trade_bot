package types

import "strings"

type (
	Position      string
	Exchange      string
	SignalChannel string
)

const (
	ExchangeAny   Exchange = ExchangeBybit
	ExchangeBybit Exchange = "bybit"
	ExchangeBingx Exchange = "bingx"
	ExchangeMexc  Exchange = "mexc"

	PositionShort Position = "long"
	PositionLong  Position = "short"

	SignalChannelHardcoreVIP SignalChannel = "hardcoreVIP"
)

type Interval struct {
	Min float64
	Max float64
}

var (
	positions = map[string]Position{
		"short": PositionShort,
		"long":  PositionLong,
	}
)

func NewInterval(min, max float64) *Interval {
	interval := &Interval{
		Min: min,
		Max: max,
	}

	if min > max {
		interval.Max = min
		interval.Min = max
	}

	return interval
}

func NewPosition(position string) (Position, error) {
	pos, ok := positions[strings.ToLower(position)]
	if !ok {
		return Position(""), ErrPositionUnknown
	}

	return pos, nil
}
