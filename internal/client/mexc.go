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

	"trade_bot/internal/client/types"
	commonTypes "trade_bot/internal/types"
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
	mexcOrderType = map[commonTypes.OrderType]string{
		commonTypes.OrderTypeLimit:             "LIMIT",
		commonTypes.OrderTypeMarket:            "MARKET",
		commonTypes.OrderTypeLimitMarket:       "LIMIT_MARKET",
		commonTypes.OrderTypeImmediateOrCancel: "IMMEDIATE_OR_CANCEL",
		commonTypes.OrderTypeFillOrKill:        "FILL_OR_KILL",
	}

	mexcOrderPosition = map[commonTypes.Position]string{
		commonTypes.PositionLong:  "BUY",
		commonTypes.PositionShort: "SELL",
	}
)

var (
	ErrMexcIntervalNotFound    = errors.New("mexc interval not found")
	ErrMexcOrderSideNotFound   = errors.New("mexc order side not found")
	ErrMexcOrderTypeNotFound   = errors.New("mexc order type not found")
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

func (m *Mexc) CreateSpotOrder(ctx context.Context, order *types.SpotOrder) error {
	orderPosition, ok := mexcOrderPosition[order.Position]
	if !ok {
		return ErrMexcOrderSideNotFound
	}

	orderType, ok := mexcOrderType[order.Type]
	if !ok {
		return ErrMexcOrderTypeNotFound
	}

	queryParams := url.Values{}
	queryParams.Set("symbol", fmt.Sprintf("%s%s", order.Symbol, order.BaseSymbol))
	queryParams.Set("side", orderPosition)
	queryParams.Set("type", orderType)
	queryParams.Set("quantity", strconv.FormatFloat(order.Quantity, 'f', 6, 64))
	queryParams.Set("price", strconv.FormatFloat(order.Entry, 'f', 6, 64))

	bytes, err := m.doRequest(ctx, http.MethodPost, "/api/v3/order", queryParams)
	if err != nil {
		return fmt.Errorf("Mexc::CreateOrder : %w", err)
	}

	var orderRecv orderCreated
	err = json.Unmarshal(bytes, &orderRecv)
	if err != nil {
		return fmt.Errorf("Mexc::CreateOrder : %w", err)
	}

	return nil
}

func (m *Mexc) CancelAllOrders(ctx context.Context, symbol, baseSymbol string) error {
	currency := fmt.Sprintf("%s%s", symbol, baseSymbol)

	queryParams := url.Values{}
	queryParams.Set("symbol", currency)

	_, err := m.doRequest(ctx, http.MethodDelete, "/api/v3/openOrders", queryParams)
	if err != nil {
		return fmt.Errorf("CancelAllOrders : %w", err)
	}

	return nil
}

func (m *Mexc) GetPrice(ctx context.Context, symbol, baseSymbol string) (float64, error) {
	currency := fmt.Sprintf("%s%s", symbol, baseSymbol)

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

func (m *Mexc) GetAssets(ctx context.Context, symbol string) (float64, error) {
	bytes, err := m.doRequest(ctx, http.MethodGet, "/api/v3/account", url.Values{})
	if err != nil {
		return 0, fmt.Errorf("Mexc::GetAssets : %w", err)
	}

	var account accountMexc
	err = json.Unmarshal(bytes, &account)
	if err != nil {
		return 0, fmt.Errorf("Mexc::GetAssets : %w", err)
	}

	for _, balance := range account.Balances {
		if balance.Currency == symbol {
			floatValue, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return 0, fmt.Errorf("Mexc::GetAssets : %w", err)
			}

			return floatValue, nil
		}
	}

	return 0, ErrAssetNotFound
}
