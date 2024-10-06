package repository

import (
	"context"
	"errors"
	"sync"

	"trade_bot/internal/types"
)

type Asset struct {
	idxCurrency map[string]types.AssetEntity

	mu sync.RWMutex
}

var ErrAssetNotFound = errors.New("asset entry is not found")

func NewAsset() *Asset {
	return &Asset{
		idxCurrency: make(map[string]types.AssetEntity),
	}
}

func (a *Asset) Exist(ctx context.Context, currency string) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	_, ok := a.idxCurrency[currency]
	return ok, nil
}

func (a *Asset) Add(ctx context.Context, entity *types.AssetEntity) error {
	return a.save(ctx, entity)
}

func (a *Asset) Update(ctx context.Context, entity *types.AssetEntity) error {
	return a.save(ctx, entity)
}

func (a *Asset) save(_ context.Context, entity *types.AssetEntity) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.idxCurrency[entity.Currency] = *entity

	return nil
}
