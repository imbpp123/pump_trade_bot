package domain

import (
	"context"
	"fmt"
	"time"

	"trade_bot/internal/types"
)

type assetClient interface {
	GetAssets(ctx context.Context) ([]types.Asset, error)
}

type assetRepository interface {
	Exist(ctx context.Context, currency string) (bool, error)
	Add(ctx context.Context, entity *types.AssetEntity) error
	Update(ctx context.Context, entity *types.AssetEntity) error
}

type Asset struct {
	client     assetClient
	repository assetRepository
}

func NewAsset(
	client assetClient,
	assetRepository assetRepository,
) *Asset {
	return &Asset{
		client:     client,
		repository: assetRepository,
	}
}

func (a *Asset) updateAssets(ctx context.Context) error {
	assets, err := a.client.GetAssets(ctx)
	if err != nil {
		return fmt.Errorf("domain.Asset.updateAssets : %w", err)
	}

	for _, asset := range assets {
		exists, err := a.repository.Exist(ctx, asset.Currency)
		if err != nil {
			return fmt.Errorf("domain.Asset.updateAssets : %w", err)
		}

		newEntity := &types.AssetEntity{
			Currency: asset.Currency,
			Amount:   asset.Amount,
		}
		if !exists {
			err := a.repository.Add(ctx, newEntity)
			if err != nil {
				return fmt.Errorf("domain.Asset.updateAssets : %w", err)
			}
		} else {
			err := a.repository.Update(ctx, newEntity)
			if err != nil {
				return fmt.Errorf("domain.Asset.updateAssets : %w", err)
			}
		}
	}

	fmt.Println(a.repository)

	return nil
}

func (a *Asset) RunUpdate(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-ticker.C:
			if err := a.updateAssets(ctx); err != nil {
				return fmt.Errorf("domain.Asset.RunUpdate : %w", err)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
