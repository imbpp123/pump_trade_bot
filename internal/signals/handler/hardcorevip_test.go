package handler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	chatTypes "trade_bot/internal/chat/types"
	"trade_bot/internal/signals/handler"
	"trade_bot/internal/signals/types"
)

func TestHardcoreVIPParse(t *testing.T) {
	text := `📈 LONG
 
	▪️Монета: ETC
	▪️Плечо: 25-50х
	▪️Вход: от 18.715 до 18.154
	▪️Цель: 18.902
	▪️Стоп: 17.592`

	handler := handler.NewHardcoreVIP()

	signal, err := handler.ParseSignal(context.Background(), &chatTypes.ChatIncomingMessage{
		Text: text,
	})
	assert.NoError(t, err)

	assert.NotNil(t, signal)
	assert.Equal(
		t,
		&types.Signal{
			Symbol:    "ETC",
			Direction: types.LongDirection,
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

/*
func TestHardcoreVIPGetEntryPoints(t *testing.T) {
	type testCase struct {
		pointQTY    int
		direction   types.Direction
		price       float64
		entryPeriod types.Period
		expected    []float64
	}

	testCases := map[string]testCase{
		"0 point QTY": {
			pointQTY:  0,
			direction: types.LongDirection,
			price:     10,
			entryPeriod: types.Period{
				Min: 1,
				Max: 101,
			},
			expected: nil,
		},
		"1 point QTY": {
			pointQTY:  1,
			direction: types.LongDirection,
			price:     120,
			entryPeriod: types.Period{
				Min: 1,
				Max: 101,
			},
			expected: []float64{51},
		},
		"1 point QTY price less max - long": {
			pointQTY:  1,
			direction: types.LongDirection,
			price:     101,
			entryPeriod: types.Period{
				Min: 1,
				Max: 111,
			},
			expected: []float64{51},
		},
		"1 point QTY price greater min - short": {
			pointQTY:  1,
			direction: types.ShortDirection,
			price:     11,
			entryPeriod: types.Period{
				Min: 1,
				Max: 111,
			},
			expected: []float64{61},
		},
		"2 point QTY": {
			pointQTY:  2,
			direction: types.LongDirection,
			price:     120,
			entryPeriod: types.Period{
				Min: 1,
				Max: 101,
			},
			expected: []float64{11, 91},
		},
		"3 point QTY": {
			pointQTY:  3,
			direction: types.LongDirection,
			price:     120,
			entryPeriod: types.Period{
				Min: 1,
				Max: 101,
			},
			expected: []float64{11, 51, 91},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			handler := handler.NewHardcoreVIP()

			entries := handler.GetEntryPoints(tc.pointQTY, tc.direction, tc.price, &tc.entryPeriod)
			assert.Equal(t, tc.expected, entries)
		})
	}
}
*/
