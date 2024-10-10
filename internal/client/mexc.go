package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"trade_bot/internal/types"
)

const (
	baseURL         string = "https://api.mexc.com"
	mexcCandleLimit int    = 10
)

type currencyPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type orderCreated struct {
	Currency string `json:"symbol"`
	OrderID  string `json:"orderId"`
	Price    string `json:"price"`
}

type balanceMexc struct {
	Currency string `json:"asset"`
	Free     string `json:"free"`
	Locked   string `json:"locked"`
}

type accountMexc struct {
	Balances []balanceMexc `json:"balances"`
}

type Mexc struct {
	apiKey    string
	secretKey string
	baseUrl   string
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

	mexcOrderType = map[types.OrderType]string{
		types.OrderTypeLimit:             "LIMIT",
		types.OrderTypeMarket:            "MARKET",
		types.OrderTypeLimitMarket:       "LIMIT_MARKET",
		types.OrderTypeImmediateOrCancel: "IMMEDIATE_OR_CANCEL",
		types.OrderTypeFillOrKill:        "FILL_OR_KILL",
	}

	mexcOrderSide = map[types.OrderSide]string{
		types.OrderSideLong:  "BUY",
		types.OrderSideShort: "SELL",
	}
)

var (
	ErrMexcIntervalNotFound    = errors.New("mexc interval not found")
	ErrMexcOrderSideNotFound   = errors.New("mexc order side not found")
	ErrMexcOrderTypeNotFound   = errors.New("mexc order type not found")
	ErrMexcOrderStatusNotFound = errors.New("mexc order status not found")
	ErrAssetNotFound           = errors.New("asset not found")
)

func NewMexc(
	apiKey, apiSecret string,
) *Mexc {
	return &Mexc{
		apiKey:    apiKey,
		secretKey: apiSecret,
		baseUrl:   baseURL,
	}
}

func (m *Mexc) CreateOrder(ctx context.Context, order *types.OrderCreate) (types.OrderID, error) {
	orderSide, ok := mexcOrderSide[order.Side]
	if !ok {
		return "", ErrMexcOrderSideNotFound
	}

	orderType, ok := mexcOrderType[order.Type]
	if !ok {
		return "", ErrMexcOrderTypeNotFound
	}

	queryParams := url.Values{}
	queryParams.Set("symbol", order.Currency)
	queryParams.Set("side", orderSide)
	queryParams.Set("type", orderType)
	queryParams.Set("quantity", strconv.FormatFloat(order.Quantity, 'f', 6, 64))
	queryParams.Set("price", strconv.FormatFloat(order.Price, 'f', 6, 64))

	bytes, err := m.doRequest(ctx, http.MethodPost, "/api/v3/order", queryParams)
	if err != nil {
		return "", fmt.Errorf("CreateOrder : %w", err)
	}

	var orderRecv orderCreated
	err = json.Unmarshal(bytes, &orderRecv)
	if err != nil {
		return "", fmt.Errorf("CreateOrder : %w", err)
	}

	return types.OrderID(orderRecv.OrderID), nil
}

func (m *Mexc) CancelAllOrders(ctx context.Context, currency string) error {
	queryParams := url.Values{}
	queryParams.Set("symbol", currency)

	_, err := m.doRequest(ctx, http.MethodDelete, "/api/v3/openOrders", queryParams)
	if err != nil {
		return fmt.Errorf("CancelAllOrders : %w", err)
	}

	return nil
}

func (m *Mexc) GetCurrencyPriceTicker(ctx context.Context, currency string) (float64, error) {
	queryParams := url.Values{}
	queryParams.Set("symbol", currency)

	bytes, err := m.doRequest(ctx, http.MethodGet, "/api/v3/ticker/price", queryParams)
	if err != nil {
		return 0, fmt.Errorf("GetCurrencyPriceTicker : %w", err)
	}

	var price currencyPrice
	err = json.Unmarshal(bytes, &price)
	if err != nil {
		return 0, fmt.Errorf("GetCurrencyPriceTicker : %w", err)
	}

	floatValue, err := strconv.ParseFloat(price.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse price to float: %w", err)
	}

	return floatValue, nil
}

func (m *Mexc) doRequest(_ context.Context, method, url string, queryParams url.Values) ([]byte, error) {
	queryParams.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	mac := hmac.New(sha256.New, []byte(m.secretKey))
	mac.Write([]byte(queryParams.Encode()))

	queryParams.Set("signature", hex.EncodeToString(mac.Sum(nil)))

	requestURL := fmt.Sprintf("%s%s?%s", baseURL, url, queryParams.Encode())

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("doRequest : %w", err)
	}

	req.Header.Add("X-MEXC-APIKEY", m.apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response : %w", err)
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		return nil, fmt.Errorf("received error status %d with message: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (m *Mexc) GetCurrencyCandles(ctx context.Context, currency string, interval types.CandleInterval) ([]types.Candle, error) {
	intervalStr, ok := mexcIntervals[interval]
	if !ok {
		return nil, ErrMexcIntervalNotFound
	}

	queryParams := url.Values{}
	queryParams.Set("symbol", currency)
	queryParams.Set("interval", intervalStr)
	queryParams.Set("limit", strconv.Itoa(mexcCandleLimit)) // don't need default 500 candles..

	bytes, err := m.doRequest(ctx, http.MethodGet, "/api/v3/klines", queryParams)
	if err != nil {
		return nil, fmt.Errorf("GetCurrencyCandles : %w", err)
	}

	var candles [][]interface{}
	err = json.Unmarshal(bytes, &candles)
	if err != nil {
		return nil, fmt.Errorf("GetCurrencyCandles : %w", err)
	}

	var result []types.Candle

	for _, kline := range candles {
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

func (m *Mexc) GetAssets(ctx context.Context, currency string) (float64, error) {
	bytes, err := m.doRequest(ctx, http.MethodGet, "/api/v3/account", url.Values{})
	if err != nil {
		return 0, fmt.Errorf("CreateOrder : %w", err)
	}

	var account accountMexc
	err = json.Unmarshal(bytes, &account)
	if err != nil {
		return 0, fmt.Errorf("GetAssets : %w", err)
	}

	for _, balance := range account.Balances {
		if balance.Currency == currency {
			floatValue, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return 0, fmt.Errorf("GetAssets : %w", err)
			}

			return floatValue, nil
		}
	}

	return 0, ErrAssetNotFound
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
