package types

const (
	ChatTypeTelegram int = 1
)

type ChatIncomingMessage struct {
	ChatType int    `json:"chat_type"`
	Text     string `json:"text"`
	ChatID   string `json:"chat_id"`
	UserID   string `json:"user_id"`
}
