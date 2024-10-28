package handler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	chatTypes "trade_bot/internal/chat/types"
	"trade_bot/internal/signals/types"
)

const (
	hardcoreVIPBaseSymbol  string = "USDT"
	hardcoreVIPEntryPoints int    = 3
	hardcoreVIPChatID      string = "1566432615"
)

type HardcoreVIP struct {
}

func NewHardcoreVIP() *HardcoreVIP {
	return &HardcoreVIP{}
}

func (h *HardcoreVIP) Name() string {
	return "HardcodeVIP"
}

func (h *HardcoreVIP) CanHandle(ctx context.Context, message *chatTypes.ChatIncomingMessage) bool {
	return message.ChatID == hardcoreVIPChatID
}

func (h *HardcoreVIP) ParseSignal(ctx context.Context, chatMessage *chatTypes.ChatIncomingMessage) (*types.Signal, error) {
	var (
		signal types.Signal
		err    error
	)

	message := chatMessage.Text

	// direction
	directionPattern := regexp.MustCompile(`📈\s*(LONG|SHORT)\n`)
	directionMatches := directionPattern.FindStringSubmatch(message)
	if len(directionMatches) == 0 {
		return nil, types.ErrParseDirectionNotFound

	}
	signal.Direction = types.Direction(strings.ToLower(directionMatches[1]))

	// symbol
	symbolPattern := regexp.MustCompile(`Монета:\s*(\w+)\n`)
	symbolMatches := symbolPattern.FindStringSubmatch(message)
	if len(symbolMatches) == 0 {
		return nil, types.ErrParseSymbolNotFound
	}
	signal.Symbol = symbolMatches[1]

	// leverage
	leveragePattern := regexp.MustCompile(`Плечо:\s*([\d]+)-([\d]+)х\n`)
	leverageMatches := leveragePattern.FindStringSubmatch(message)
	if len(leverageMatches) == 0 {
		return nil, types.ErrParseLeverageNotFound
	}
	signal.Leverage.Min, err = strconv.ParseFloat(leverageMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	signal.Leverage.Max, err = strconv.ParseFloat(leverageMatches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}

	// entry
	entryPattern := regexp.MustCompile(`Вход:\s*от\s*([\d.]+)\s*до\s*([\d.]+)\n`)
	entryMatches := entryPattern.FindStringSubmatch(message)
	if len(entryMatches) != 3 {
		return nil, types.ErrParseEntryNotFound
	}
	signal.Entry.Min, err = strconv.ParseFloat(entryMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	signal.Entry.Max, err = strconv.ParseFloat(entryMatches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}

	// target
	targetPattern := regexp.MustCompile(`Цель:\s*([\d.]+)\n`)
	targetMatches := targetPattern.FindStringSubmatch(message)
	if len(targetMatches) != 2 {
		return nil, types.ErrParseTargetNotFound
	}
	signal.Target.Min, err = strconv.ParseFloat(targetMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	signal.Target.Max = signal.Target.Min

	// stop
	stopPattern := regexp.MustCompile(`Стоп:\s*([\d.]+)`)
	stopMatches := stopPattern.FindStringSubmatch(message)
	if len(stopMatches) != 2 {
		return nil, types.ErrParseStopNotFound
	}
	signal.Stop, err = strconv.ParseFloat(stopMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}

	return &signal, nil
}

/*
func (h *HardcoreVIP) ProcessSignal(ctx context.Context, signal *types.Signal) error {
	currency := signal.Symbol + hardcoreVIPBaseSymbol

	// get price for symbol
	price, err := h.client.GetPrice(ctx, currency)
	if err != nil {
		return fmt.Errorf("HardcoreVIP::ProcessSignal : %w", err)
	}

	// create orders
	orders := []types.Order

	// get entry points
	entryPoints := h.GetEntryPoints(hardcoreVIPEntryPoints, signal.Direction, price, &signal.Entry)

	// fill entries to orders

	fmt.Println(entryPoints)

	// send request to exchange to create orders

	return nil
}

func (h *HardcoreVIP) GetEntryPoints(pointQTY int, direction types.Direction, price float64, entryPeriod *types.Period) []float64 {
	var entries []float64

	if pointQTY == 0 {
		return nil
	}

	min, max := entryPeriod.Min, entryPeriod.Max
	if direction == types.LongDirection && price < max {
		max = price
	}
	if direction == types.ShortDirection && price > min {
		min = price
	}
	if pointQTY == 1 {
		return []float64{(max + min) / 2}
	}

	distance := max - min
	max -= distance / 10 // minus 10% from max
	min += distance / 10 // plus 10% to min

	step := (max - min) / float64(pointQTY-1)
	for i := 0; i < pointQTY; i++ {
		entries = append(entries, min+float64(i)*step)
	}

	return entries
}
*/
