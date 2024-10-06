package types

import "time"

type CandleInterval int

const (
	CandleInterval1m CandleInterval = iota
	CandleInterval5m
	CandleInterval15m
	CandleInterval30m
	CandleInterval1h
	CandleInterval4h
	CandleInterval1d
	CandleInterval1W
	CandleInterval1M
)

type Candle struct {
	OpenTime    time.Time
	CloseTime   time.Time
	Open        float64
	High        float64
	Low         float64
	Close       float64
	Volume      float64
	AssetVolume float64
	Interval    CandleInterval
}
