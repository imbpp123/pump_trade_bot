package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChatTypeTelegram int = 1
)

type ChatIncomingMessage struct {
	UUID      uuid.UUID `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	ChatType  int       `json:"chat_type"`
	Text      string    `json:"text"`
	ChatID    string    `json:"chat_id"`
}

func NewChatIncomingMessage(
	chatType int,
	text string,
	chatID string,
) *ChatIncomingMessage {
	return &ChatIncomingMessage{
		UUID:      uuid.New(),
		CreatedAt: time.Now(),
		ChatType:  chatType,
		Text:      text,
		ChatID:    chatID,
	}
}
