package signals

import (
	"context"
	"fmt"

	"trade_bot/internal/signals/types"
)

type Handler interface {
	CanHandle(ctx context.Context, chatID string) bool
	ParseSignal(ctx context.Context, message string) (*types.Signal, error)
	ProcessSignal(ctx context.Context, signal *types.Signal)
}

type Message struct {
	ChatID  string
	Message string
}

type MessageController struct {
	handlers []Handler
}

func NewMessageProcessor(handlers []Handler) *MessageController {
	return &MessageController{
		handlers: handlers,
	}
}

func (m *MessageController) ProcessMessage(ctx context.Context, message *Message) error {
	for _, h := range m.handlers {
		if h.CanHandle(ctx, message.ChatID) {
			signal, err := h.ParseSignal(ctx, message.Message)
			if err != nil {
				return fmt.Errorf("MessageController::ProcessMessage : %w", err)
			}

			go h.ProcessSignal(ctx, signal)
		}
	}

	return nil
}
