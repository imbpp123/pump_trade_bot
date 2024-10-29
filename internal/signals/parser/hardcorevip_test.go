package parser_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	chatTypes "trade_bot/internal/chat/types"
	"trade_bot/internal/signals/parser"
	commonTypes "trade_bot/internal/types"
)

func TestHardcoreVIPParseLong(t *testing.T) {
	text := `📈 LONG
 
	▪️Монета: ETC
	▪️Плечо: 25-50х
	▪️Вход: от 18.715 до 18.154
	▪️Цель: 18.902
	▪️Стоп: 17.592`

	handler := parser.NewHardcoreVIP()

	signal, err := handler.ParseSignal(context.Background(), &chatTypes.ChatIncomingMessage{
		Text: text,
	})
	assert.NoError(t, err)

	assert.NotNil(t, signal)
	assert.Equal(t, "ETC", signal.Symbol)
	assert.Equal(t, commonTypes.PositionLong, signal.Position)

	if assert.NotNil(t, signal.LeverageInterval) {
		assert.Equal(t, 25.0, signal.LeverageInterval.Min)
		assert.Equal(t, 50.0, signal.LeverageInterval.Max)
	}

	if assert.NotNil(t, signal.EntryInterval) {
		assert.Equal(t, 18.154, signal.EntryInterval.Min)
		assert.Equal(t, 18.715, signal.EntryInterval.Max)
	}

	assert.Equal(t, 18.902, signal.Target)
	assert.Equal(t, 17.592, signal.Stop)
}

func TestHardcoreVIPParseShort(t *testing.T) {
	text := `📈 SHORT
 
	▪️Монета: ETC
	▪️Плечо: 25-50х
	▪️Вход: от 18.715 до 18.154
	▪️Цель: 17.902
	▪️Стоп: 18.592`

	handler := parser.NewHardcoreVIP()

	signal, err := handler.ParseSignal(context.Background(), &chatTypes.ChatIncomingMessage{
		Text: text,
	})
	assert.NoError(t, err)

	assert.NotNil(t, signal)
	assert.Equal(t, "ETC", signal.Symbol)
	assert.Equal(t, commonTypes.PositionShort, signal.Position)

	if assert.NotNil(t, signal.LeverageInterval) {
		assert.Equal(t, 25.0, signal.LeverageInterval.Min)
		assert.Equal(t, 50.0, signal.LeverageInterval.Max)
	}

	if assert.NotNil(t, signal.EntryInterval) {
		assert.Equal(t, 18.154, signal.EntryInterval.Min)
		assert.Equal(t, 18.715, signal.EntryInterval.Max)
	}

	assert.Equal(t, 17.902, signal.Target)
	assert.Equal(t, 18.592, signal.Stop)
}

func TestHardcoreVIPParsePosition(t *testing.T) {
	text := "📈 LONG \n \n▪️Монета: SOL\n▪️Плечо: 25-50х\n▪️Вход: от 181.42 до 175.98\n▪️Цель: 183.23\n▪️Стоп: 170.53"

	handler := parser.NewHardcoreVIP()

	signal, err := handler.ParseSignal(context.Background(), &chatTypes.ChatIncomingMessage{
		Text: text,
	})
	assert.NoError(t, err)

	assert.NotNil(t, signal)
	assert.Equal(t, "SOL", signal.Symbol)
	assert.Equal(t, commonTypes.PositionLong, signal.Position)

	if assert.NotNil(t, signal.LeverageInterval) {
		assert.Equal(t, 25.0, signal.LeverageInterval.Min)
		assert.Equal(t, 50.0, signal.LeverageInterval.Max)
	}

	if assert.NotNil(t, signal.EntryInterval) {
		assert.Equal(t, 175.98, signal.EntryInterval.Min)
		assert.Equal(t, 181.42, signal.EntryInterval.Max)
	}

	assert.Equal(t, 183.23, signal.Target)
	assert.Equal(t, 170.53, signal.Stop)
}
