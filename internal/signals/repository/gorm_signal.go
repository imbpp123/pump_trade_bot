package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"trade_bot/internal/signals/types"
	commonTypes "trade_bot/internal/types"
)

type gormSignalEntity struct {
	UUID                 uuid.UUID `gorm:"primaryKey"`
	CreatedAt            time.Time
	Exchange             commonTypes.Exchange
	Channel              commonTypes.SignalChannel
	Symbol               string
	BaseSymbol           string
	Position             commonTypes.Position
	LeverageIntervalFrom *float64
	LeverageIntervalTo   *float64
	EntryIntervalFrom    *float64
	EntryIntervalTo      *float64
	Target               float64
	Stop                 float64
}

func newEntityFromSignal(signal *types.Signal) *gormSignalEntity {
	entity := &gormSignalEntity{
		UUID:       signal.UUID,
		CreatedAt:  signal.CreatedAt,
		Exchange:   signal.Exchange,
		Channel:    signal.Channel,
		Symbol:     signal.Symbol,
		BaseSymbol: signal.BaseSymbol,
		Position:   signal.Position,
		Target:     signal.Target,
		Stop:       signal.Stop,
	}

	if signal.EntryInterval != nil {
		entity.EntryIntervalFrom = &signal.EntryInterval.Min
		entity.EntryIntervalTo = &signal.EntryInterval.Max
	}

	if signal.LeverageInterval != nil {
		entity.LeverageIntervalFrom = &signal.LeverageInterval.Min
		entity.LeverageIntervalTo = &signal.LeverageInterval.Max
	}

	return entity
}

type GormSignal struct {
	db *gorm.DB
}

func NewGormSignal(
	db *gorm.DB,
) (*GormSignal, error) {
	if err := db.AutoMigrate(&gormSignalEntity{}); err != nil {
		return nil, fmt.Errorf("NewGormSignal : %w", err)
	}

	return &GormSignal{
		db: db,
	}, nil
}

func (g *GormSignal) Create(ctx context.Context, signal *types.Signal) error {
	if err := g.db.WithContext(ctx).Create(newEntityFromSignal(signal)).Error; err != nil {
		return fmt.Errorf("GormSignal::Create : %w", err)
	}

	return nil
}
