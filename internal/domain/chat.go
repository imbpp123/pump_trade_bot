package domain

import (
	"context"
	"fmt"
	"strings"

	"trade_bot/internal/types"
)

type Chat struct{}

func NewChat() *Chat {
	return &Chat{}
}

func (c *Chat) WaitForPumpCurrency(ctx context.Context, channel types.ChannelMessageIn) (string, error) {
	searchString := "Next message is the coin name"
	var prevMsg *types.ChatMessageIn

	fmt.Println("Wait for message")

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case msg := <-channel:
			if msg == nil {
				continue
			}

			if msg != nil && msg.ChatID != nil {
				fmt.Printf("chat message: %+v chat_id = %s\n", msg, *msg.ChatID)
			} else {
				fmt.Printf("chat message: %+v\n", msg)
			}

			if prevMsg == nil {
				if strings.Contains(msg.Text, searchString) {
					fmt.Printf("PRE chat message: %+v\n", msg)
					prevMsg = msg
					break
				}
			} else if prevMsg.ChatID != nil && msg.ChatID != nil && *prevMsg.ChatID == *msg.ChatID {
				fmt.Printf("found!\n")
				return msg.Text, nil
			}
		}
	}
}
