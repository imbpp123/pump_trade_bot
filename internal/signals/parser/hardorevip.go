package parser

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	chatTypes "trade_bot/internal/chat/types"
	"trade_bot/internal/signals/types"
	commonTypes "trade_bot/internal/types"
)

const (
	hardcoreVIPChatID string = "1566432615"
)

type HardcoreVIP struct {
}

func NewHardcoreVIP() *HardcoreVIP {
	return &HardcoreVIP{}
}

func (h *HardcoreVIP) Name() commonTypes.SignalChannel {
	return commonTypes.SignalChannelHardcoreVIP
}

func (h *HardcoreVIP) CanHandle(ctx context.Context, message *chatTypes.ChatIncomingMessage) bool {
	return message.ChatID == hardcoreVIPChatID
}

func (h *HardcoreVIP) ParseSignal(ctx context.Context, chatMessage *chatTypes.ChatIncomingMessage) (*types.Signal, error) {
	var (
		err error
	)

	message := chatMessage.Text
	signal := types.NewSignal()
	signal.Channel = h.Name()
	signal.Exchange = commonTypes.ExchangeBybit

	// position
	positionPattern := regexp.MustCompile(`📈\s*(LONG|SHORT)\n`)
	positionMatches := positionPattern.FindStringSubmatch(message)
	if len(positionMatches) == 0 {
		return nil, types.ErrParsePositionNotFound

	}
	signal.Position, err = commonTypes.NewPosition(positionMatches[1])
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}

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
	leverageMin, err := strconv.ParseFloat(leverageMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	leverageMax, err := strconv.ParseFloat(leverageMatches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	signal.LeverageInterval = commonTypes.NewInterval(leverageMin, leverageMax)

	// entry
	entryPattern := regexp.MustCompile(`Вход:\s*от\s*([\d.]+)\s*до\s*([\d.]+)\n`)
	entryMatches := entryPattern.FindStringSubmatch(message)
	if len(entryMatches) != 3 {
		return nil, types.ErrParseEntryNotFound
	}
	entryMin, err := strconv.ParseFloat(entryMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	entryMax, err := strconv.ParseFloat(entryMatches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}
	signal.EntryInterval = commonTypes.NewInterval(entryMin, entryMax)

	// target
	targetPattern := regexp.MustCompile(`Цель:\s*([\d.]+)\n`)
	targetMatches := targetPattern.FindStringSubmatch(message)
	if len(targetMatches) != 2 {
		return nil, types.ErrParseTargetNotFound
	}
	signal.Target, err = strconv.ParseFloat(targetMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("HardcoreVIP::ParseSignal : %w", err)
	}

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

	return signal, nil
}
