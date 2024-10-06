package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"mexc-sdk/mexcsdk"

	"trade_bot/internal/types"
)

type Mexc struct {
	clientSpot   mexcsdk.Spot
	clientMarket mexcsdk.Market
}

var mexcIntervals = map[types.CandleInterval]string{
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

var ErrMexcIntervalNotFound = errors.New("mexc interval not found")

func NewMexc(
	apiKey, apiSecret string,
) *Mexc {
	clientSpot := mexcsdk.NewSpot(&apiKey, &apiSecret)
	clientMarket := mexcsdk.NewMarket(&apiKey, &apiSecret)

	return &Mexc{
		clientSpot:   clientSpot,
		clientMarket: clientMarket,
	}
}

func (m *Mexc) GetCurrencyPriceTicker(ctx context.Context, currency string) (*types.CurrencyPrice, error) {
	info := m.clientMarket.TickerPrice(&currency)

	data, ok := info.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast response: %v", info)
	}

	currencyStr, ok := data["symbol"].(string)
	if !ok {
		return nil, fmt.Errorf("can't cast symbol: %v", data)
	}

	price, err := anyStringToFloat64(data["price"])
	if err != nil {
		return nil, fmt.Errorf("can't parse price: %w", err)
	}

	return &types.CurrencyPrice{
		Currency: currencyStr,
		Price:    price,
	}, nil
}

func (m *Mexc) GetCurrencyCandles(ctx context.Context, currency string, interval types.CandleInterval) ([]types.Candle, error) {
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

func (m *Mexc) GetAssets(ctx context.Context) ([]types.Asset, error) {
	info := m.clientSpot.AccountInfo()

	data, ok := info.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast response: %v", info)
	}

	balances, ok := data["balances"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast balances: %v", data["balances"])
	}

	var result []types.Asset

	for _, balance := range balances {
		balanceInfo, ok := balance.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("can't case balance entry: %v", balance)
		}

		asset, ok := balanceInfo["asset"].(string)
		if !ok {
			return nil, fmt.Errorf("can't case asset: %v", balanceInfo["asset"])
		}

		amountStr, ok := balanceInfo["free"].(string)
		if !ok {
			return nil, fmt.Errorf("can't case amount: %v", balanceInfo["free"])
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("can't convert amount: %w", err)
		}

		result = append(result, types.Asset{
			Currency: asset,
			Amount:   amount,
		})
	}

	return result, nil
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
