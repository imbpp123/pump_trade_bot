package types

type OrderSide int

const (
	OrderSideLong OrderSide = iota
	OrderSideShort
)

type OrderID string


type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota
	OrderStatusFilled
	OrderStatusPartiallyFilled
	OrderStatusCanceled
	OrderStatusPartiallyCanceled
)

type OrderCreate struct {
	Currency string
	Side     OrderSide
	Type     OrderType
	Quantity float64
	Price    float64
}

type Order struct {
	OrderID  string
	Currency string
	Side     OrderSide
	Type     OrderType
	Quantity float64
	Price    float64
	Status   OrderStatus
}
