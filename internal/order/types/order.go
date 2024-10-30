package types

import (
	"time"

	"github.com/google/uuid"

	commonTypes "trade_bot/internal/types"
)

type OrderStatus string

const OrderStatusNew OrderStatus = "new"

type Order struct {
	UUID       uuid.UUID
	SignalUUID uuid.UUID
	CreatedAt  time.Time
	Exchange   commonTypes.Exchange

	Symbol     string
	BaseSymbol string
	Position   commonTypes.Position

	Leverage float64
	Entry    float64
	Quantity float64

	Target float64
	Stop   float64

	Status OrderStatus
}
