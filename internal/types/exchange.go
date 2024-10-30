package types

import "strings"

type (
	Position      string
	Exchange      string
	SignalChannel string

	OrderType string
)

const (
	ExchangeAny   Exchange = ExchangeBybit
	ExchangeBybit Exchange = "bybit"
	ExchangeBingx Exchange = "bingx"
	ExchangeMexc  Exchange = "mexc"

	PositionShort Position = "long"
	PositionLong  Position = "short"

	SignalChannelHardcoreVIP SignalChannel = "hardcoreVIP"

	OrderTypeLimit             OrderType = "limit"
	OrderTypeMarket            OrderType = "market"
	OrderTypeLimitMarket       OrderType = "limit_market"
	OrderTypeImmediateOrCancel OrderType = "immediate_or_cancel"
	OrderTypeFillOrKill        OrderType = "fill_or_kill"
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
