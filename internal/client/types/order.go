package types

import (
	commonTypes "trade_bot/internal/types"
)

type SpotOrder struct {
	Exchange   commonTypes.Exchange
	Type       commonTypes.OrderType
	Position   commonTypes.Position
	Symbol     string
	BaseSymbol string
	Entry      float64
	Quantity   float64
}
