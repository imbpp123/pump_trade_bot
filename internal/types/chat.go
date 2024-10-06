package types

type ChannelMessageIn chan *ChatMessageIn

type ChatMessageIn struct {
	Text      string
	ChatID    *string
	UserID    *string
	IsPrivate bool
}
