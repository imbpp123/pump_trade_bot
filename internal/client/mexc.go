package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"mexc-sdk/mexcsdk"

	"trade_bot/internal/types"
)

const (
	mexcParallelRequestQTY      int           = 25
	mexcRequestWindow           time.Duration = 10 * time.Second
	mexcMaxRequestPerWindow     int           = 500
	mexcRequestCoolDownDuration time.Duration = 50 * time.Millisecond
)

type requestFunc chan func() error

type Mexc struct {
	clientSpot   mexcsdk.Spot
	clientMarket mexcsdk.Market

	requestChan requestFunc
	requests    []time.Time
	requestWg   sync.WaitGroup
	requestsMu  sync.Mutex
}

var (
	mexcIntervals = map[types.CandleInterval]string{
		types.CandleInterval1m:  "1m",
		types.CandleInterval5m:  "5m",
		types.CandleInterval15m: "15m",
		types.CandleInterval30m: "30m",
		types.CandleInterval1h:  "60m",
		types.CandleInterval4h:  "4h",
		types.CandleInterval1d:  "1d",
		types.CandleInterval1W:  "1W",
		types.CandleInterval1M:  "1M",
	}

	mexcOrderStatus = map[types.OrderStatus]string{
		types.OrderStatusNew:               "NEW",
		types.OrderStatusFilled:            "FILLED",
		types.OrderStatusPartiallyFilled:   "PARTIALLY_FILLED",
		types.OrderStatusCanceled:          "CANCELED",
		types.OrderStatusPartiallyCanceled: "PARTIALLY_CANCELED",
	}

	mexcToOrderStatus = map[string]types.OrderStatus{
		"NEW":                types.OrderStatusNew,
		"FILLED":             types.OrderStatusFilled,
		"PARTIALLY_FILLED":   types.OrderStatusPartiallyFilled,
		"CANCELED":           types.OrderStatusCanceled,
		"PARTIALLY_CANCELED": types.OrderStatusPartiallyCanceled,
	}

	mexcOrderType = map[types.OrderType]string{
		types.OrderTypeLimit:             "LIMIT",
		types.OrderTypeMarket:            "MARKET",
		types.OrderTypeLimitMarket:       "LIMIT_MARKET",
		types.OrderTypeImmediateOrCancel: "IMMEDIATE_OR_CANCEL",
		types.OrderTypeFillOrKill:        "FILL_OR_KILL",
	}

	mexcToOrderType = map[string]types.OrderType{
		"LIMIT":               types.OrderTypeLimit,
		"MARKET":              types.OrderTypeMarket,
		"LIMIT_MARKET":        types.OrderTypeLimitMarket,
		"IMMEDIATE_OR_CANCEL": types.OrderTypeImmediateOrCancel,
		"FILL_OR_KILL":        types.OrderTypeFillOrKill,
	}

	mexcOrderSide = map[types.OrderSide]string{
		types.OrderSideLong:  "BUY",
		types.OrderSideShort: "SELL",
	}

	mexcToOrderSide = map[string]types.OrderSide{
		"BUY":  types.OrderSideLong,
		"SELL": types.OrderSideShort,
	}
)

var (
	ErrMexcIntervalNotFound    = errors.New("mexc interval not found")
	ErrMexcOrderSideNotFound   = errors.New("mexc order side not found")
	ErrMexcOrderTypeNotFound   = errors.New("mexc order type not found")
	ErrMexcOrderStatusNotFound = errors.New("mexc order status not found")
)

func NewMexc(
	apiKey, apiSecret string,
) *Mexc {
	clientSpot := mexcsdk.NewSpot(&apiKey, &apiSecret)
	clientMarket := mexcsdk.NewMarket(&apiKey, &apiSecret)

	m := &Mexc{
		clientSpot:   clientSpot,
		clientMarket: clientMarket,
		requestChan:  make(requestFunc),
	}

	for i := 0; i < mexcParallelRequestQTY; i++ {
		m.requestWg.Add(1)
		go m.processRequest()
	}

	return m
}

func (m *Mexc) CreateOrder(ctx context.Context, order *types.OrderCreate) (*types.Order, error) {
	orderSide, ok := mexcOrderSide[order.Side]
	if !ok {
		return nil, ErrMexcOrderSideNotFound
	}

	orderType, ok := mexcOrderType[order.Type]
	if !ok {
		return nil, ErrMexcOrderTypeNotFound
	}

	options := make(map[string]interface{})
	options["price"] = order.Price
	options["quantity"] = order.Quantity
	options["timestamp"] = time.Now().Unix()

	response := m.clientSpot.NewOrder(&order.Currency, &orderSide, &orderType, options)

	data, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast response: %v", response)
	}

	result, err := m.fillOrder(data)
	if err != nil {
		return nil, fmt.Errorf("GetOrder : %w", err)
	}

	return result, nil
}

func (m *Mexc) CancelAllOrders(ctx context.Context, currency string) error {
	m.clientSpot.CancelOpenOrders(&currency)
	return nil
}

func (m *Mexc) GetOrder(ctx context.Context, currency, orderID string) (*types.Order, error) {
	options := make(map[string]interface{})
	options["orderId"] = orderID
	options["timestamp"] = time.Now().Unix()

	response := m.clientSpot.QueryOrder(&currency, options)

	data, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast response: %v", response)
	}

	fmt.Println(data)

	result, err := m.fillOrder(data)
	if err != nil {
		return nil, fmt.Errorf("GetOrder : %w", err)
	}

	return result, nil
}

func (m *Mexc) fillOrder(data map[string]interface{}) (*types.Order, error) {
	var (
		err error
	)

	result := &types.Order{
		Status: types.OrderStatusNew,
	}

	anyData, ok := data["status"]
	if ok {
		statusStr, ok := anyData.(string)
		if !ok {
			return nil, fmt.Errorf("can't cast status: %v", anyData)
		}
		result.Status, ok = mexcToOrderStatus[statusStr]
		if !ok {
			return nil, ErrMexcOrderStatusNotFound
		}
	}

	result.OrderID, ok = data["orderId"].(string)
	if !ok {
		return nil, fmt.Errorf("can't cast orderId: %v", data)
	}

	result.Price, err = anyStringToFloat64(data["price"])
	if err != nil {
		return nil, fmt.Errorf("CreateOrder : %w", err)
	}

	result.Currency, ok = data["symbol"].(string)
	if !ok {
		return nil, fmt.Errorf("can't cast currency: %v", data)
	}

	result.Quantity, err = anyStringToFloat64(data["origQty"])
	if err != nil {
		return nil, fmt.Errorf("CreateOrder : %w", err)
	}

	orderSideVal, ok := data["side"].(string)
	if !ok {
		return nil, fmt.Errorf("can't cast side: %v", data)
	}
	result.Side, ok = mexcToOrderSide[orderSideVal]
	if !ok {
		return nil, ErrMexcOrderSideNotFound
	}

	orderTypeVal, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("can't cast type: %v", data)
	}
	result.Type, ok = mexcToOrderType[orderTypeVal]
	if !ok {
		return nil, ErrMexcOrderTypeNotFound
	}

	return result, nil
}

func (m *Mexc) GetCurrencyPriceTicker(currency string) (float64, error) {
	info := m.clientMarket.TickerPrice(&currency)

	data, ok := info.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("can't cast response: %v", info)
	}

	price, err := anyStringToFloat64(data["price"])
	if err != nil {
		return 0, fmt.Errorf("can't parse price: %w", err)
	}

	return price, nil
}

func (m *Mexc) GetCurrencyCandles(currency string, interval types.CandleInterval) ([]types.Candle, error) {
	intervalStr, ok := mexcIntervals[interval]
	if !ok {
		return nil, ErrMexcIntervalNotFound
	}

	info := m.clientMarket.Klines(&currency, &intervalStr, nil)

	data, ok := info.([]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast response: %v", info)
	}

	var result []types.Candle

	for _, k := range data {
		kline, ok := k.([]interface{})
		if !ok {
			return nil, fmt.Errorf("can't cast kline: %v", kline)
		}

		var err error

		candle := types.Candle{
			Interval: interval,
		}

		openTimeFloat, ok := kline[0].(float64)
		if !ok {
			return nil, fmt.Errorf("can't cast open time: %v", kline)
		}
		candle.OpenTime = float64UnixToTime(openTimeFloat)

		candle.Open, err = anyStringToFloat64(kline[1])
		if err != nil {
			return nil, fmt.Errorf("can't parse open: %w", err)
		}

		candle.High, err = anyStringToFloat64(kline[2])
		if err != nil {
			return nil, fmt.Errorf("can't parse high: %w", err)
		}

		candle.Low, err = anyStringToFloat64(kline[3])
		if err != nil {
			return nil, fmt.Errorf("can't parse low: %w", err)
		}

		candle.Close, err = anyStringToFloat64(kline[4])
		if err != nil {
			return nil, fmt.Errorf("can't parse close: %w", err)
		}

		candle.Volume, err = anyStringToFloat64(kline[5])
		if err != nil {
			return nil, fmt.Errorf("can't parse volume: %w", err)
		}

		closeTimeFloat, ok := kline[6].(float64)
		if !ok {
			return nil, fmt.Errorf("can't cast close time: %v", kline)
		}
		candle.CloseTime = float64UnixToTime(closeTimeFloat)

		candle.AssetVolume, err = anyStringToFloat64(kline[7])
		if err != nil {
			return nil, fmt.Errorf("can't parse asset volume: %w", err)
		}

		result = append(result, candle)
	}

	return result, nil
}

func (m *Mexc) GetAssets(currency string) (float64, error) {
	info := m.clientSpot.AccountInfo()

	data, ok := info.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("can't cast response: %v", info)
	}

	balances, ok := data["balances"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("can't cast balances: %v", data["balances"])
	}

	var (
		result float64
		err    error
	)

	for _, balance := range balances {
		balanceInfo, ok := balance.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("can't case balance entry: %v", balance)
		}

		asset, ok := balanceInfo["asset"].(string)
		if !ok {
			return 0, fmt.Errorf("can't case asset: %v", balanceInfo["asset"])
		}
		if asset != currency {
			continue
		}

		result, err = anyStringToFloat64(balanceInfo["free"])
		if err != nil {
			return 0, fmt.Errorf("can't convert amount: %w", err)
		}
		break
	}

	return result, nil
}

func (m *Mexc) processRequest() {
	defer m.requestWg.Done()

	for req := range m.requestChan {
		m.requestsMu.Lock()

		for len(m.requests) >= mexcMaxRequestPerWindow && time.Since(m.requests[0]) < mexcRequestWindow {
			m.requestsMu.Unlock()
			time.Sleep(mexcRequestCoolDownDuration)
			m.requestsMu.Lock()
		}

		now := time.Now()
		threshold := now.Add(-mexcRequestWindow)

		newRequests := []time.Time{}
		for _, t := range m.requests {
			if t.After(threshold) {
				newRequests = append(newRequests, t)
			}
		}
		newRequests = append(newRequests, now)

		m.requests = newRequests
		m.requestsMu.Unlock()

		if err := req(); err != nil {
			fmt.Printf("Request failed: %v\n", err)
		}
	}
}

func float64UnixToTime(t float64) time.Time {
	sec := int64(t / 1000)
	nsec := int64((t - float64(sec*1000)) * 1e6)

	return time.Unix(sec, nsec)
}

func anyStringToFloat64(str interface{}) (float64, error) {
	strValue, ok := str.(string)
	if !ok {
		return 0, fmt.Errorf("can't cast str: %v", str)
	}

	floatValue, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse str to float: %w", err)
	}

	return floatValue, nil
}
