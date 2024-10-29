package types

import (
	"time"

	"github.com/google/uuid"

	commonTypes "trade_bot/internal/types"
)

type Signal struct {
	UUID             uuid.UUID
	CreatedAt        time.Time
	Exchange         commonTypes.Exchange
	Channel          commonTypes.SignalChannel
	Symbol           string
	BaseSymbol       string
	Position         commonTypes.Position
	LeverageInterval *commonTypes.Interval
	EntryInterval    *commonTypes.Interval
	Target           float64
	Stop             float64
}

func NewSignal() *Signal {
	return &Signal{
		UUID:      uuid.New(),
		CreatedAt: time.Now(),
	}
}
