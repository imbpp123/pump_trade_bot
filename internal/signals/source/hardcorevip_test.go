package source_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"trade_bot/internal/signals/source"
	"trade_bot/internal/signals/types"
)

func TestHardcoreVIPParse(t *testing.T) {
	text := `📈 LONG
 
	▪️Монета: ETC
	▪️Плечо: 25-50х
	▪️Вход: от 18.715 до 18.154
	▪️Цель: 18.902
	▪️Стоп: 17.592`

	handler := source.NewHardcoreVIP("1234", nil)

	signal, err := handler.ParseSignal(context.Background(), text)
	assert.NoError(t, err)

	assert.NotNil(t, signal)
	assert.Equal(
		t,
		&types.Signal{
			Symbol:    "ETC",
			Direction: types.LongPosition,
			Leverage: types.Period{
				Min: 25.0,
				Max: 50.0,
			},
			Entry: types.Period{
				Min: 18.715,
				Max: 18.154,
			},
			Target: types.Period{
				Min: 18.902,
				Max: 18.902,
			},
			Stop: 17.592,
		},
		signal,
	)
}
