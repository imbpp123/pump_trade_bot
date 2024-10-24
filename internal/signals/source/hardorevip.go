package source

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"trade_bot/internal/signals/types"
)

type exchangeClient interface {
}

type HardcoreVIP struct {
	chatID string
	client exchangeClient
}

func NewHardcoreVIP(
	chatID string,
	client exchangeClient,
) *HardcoreVIP {
	return &HardcoreVIP{
		chatID: chatID,
		client: client,
	}
}

func (h *HardcoreVIP) CanHandle(ctx context.Context, chatID string) bool {
	return chatID == h.chatID
}

func (h *HardcoreVIP) ParseSignal(ctx context.Context, message string) (*types.Signal, error) {
	var (
		signal types.Signal
		err    error
	)

	// direction
	directionPattern := regexp.MustCompile(`📈\s*(LONG|SHORT)\n`)
	directionMatches := directionPattern.FindStringSubmatch(message)
	if len(directionMatches) == 0 {
		return nil, types.ErrParseDirectionNotFound

	}
	signal.Direction = types.Position(strings.ToLower(directionMatches[1]))

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

func (h *HardcoreVIP) ProcessSignal(ctx context.Context, signal *types.Signal) {
	
}
