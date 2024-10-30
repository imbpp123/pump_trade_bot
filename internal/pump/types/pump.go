package types

import (
	"time"

	"github.com/google/uuid"
)

type PumpStatus string

const (
	PumpStatusNew            PumpStatus = "new"
	PumpStatusNextSymbol     PumpStatus = "next_symbol"
	PumpStatusSymbolReceived PumpStatus = "symbol_received"
)

type Pump struct {
	UUID       uuid.UUID
	CreatedAt  time.Time
	Status     PumpStatus
	Symbol     *string
	BaseSymbol *string
}

func NewPump() *Pump {
	return &Pump{
		UUID:      uuid.New(),
		CreatedAt: time.Now(),
		Status:    PumpStatusNew,
	}
}
