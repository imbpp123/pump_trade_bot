package types

type Period struct {
	Min float64
	Max float64
}

type Direction string

const (
	LongDirection  Direction = "long"
	ShortDirection Direction = "short"
)

type Signal struct {
	Symbol    string
	Direction Direction
	Leverage  Period
	Entry     Period
	Target    Period
	Stop      float64
}
