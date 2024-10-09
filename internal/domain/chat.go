package domain

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"trade_bot/internal/types"
)

const (
	pumpChatID string = "1625691880"
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

			if msg.ChatID != nil && *msg.ChatID == pumpChatID && isUppercaseAZOnly(msg.Text) {
				// this is pump channel message with crypto
				fmt.Printf("Crypto was said in chat - %s!\n", msg.Text)

				return msg.Text, nil
			}

			if prevMsg == nil {
				if strings.Contains(msg.Text, searchString) {
					fmt.Printf("PRE chat message: %+v\n", msg)
					prevMsg = msg
					break
				}
			} else if prevMsg.ChatID != nil && msg.ChatID != nil && *prevMsg.ChatID == *msg.ChatID {
				fmt.Printf("Found after PREV message!\n")
				return msg.Text, nil
			}
		}
	}
}

func isUppercaseAZOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) || r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
