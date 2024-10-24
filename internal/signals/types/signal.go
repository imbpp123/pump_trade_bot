package types

type Period struct {
	Min float64
	Max float64
}

type Position string

const (
	LongPosition  Position = "long"
	ShortPosition Position = "short"
)

type Signal struct {
	Symbol    string
	Direction Position
	Leverage  Period
	Entry     Period
	Target    Period
	Stop      float64
}
