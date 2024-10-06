package client

import (
	"context"
	"fmt"
	"strconv"

	"mexc-sdk/mexcsdk"

	"trade_bot/internal/types"
)

type Mexc struct {
	client mexcsdk.Spot
}

func NewMexc(
	apiKey, apiSecret string,
) *Mexc {
	client := mexcsdk.NewSpot(&apiKey, &apiSecret)

	return &Mexc{
		client: client,
	}
}

func (m *Mexc) GetAssets(ctx context.Context) ([]types.Asset, error) {
	info := m.client.AccountInfo()

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
