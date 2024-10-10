package domain

import (
	"context"
	"fmt"
	"regexp"
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

type parseFunc struct {
	detect func(str string) bool
	parse  func(str string) string
}

var symbolParser = map[string]parseFunc{
	"crypto_pump_signal": {
		detect: func(msg string) bool {
			return strings.Contains(msg, "The next message will be the coin name")
		},
		parse: func(msg string) string {
			re := regexp.MustCompile(`(?i)COIN NAME\s*:\s*([A-Z]+)`)

			matches := re.FindStringSubmatch(msg)
			if len(matches) > 1 {
				return strings.ToUpper(matches[1])
			}

			return ""
		},
	},
	"crypto_pump_club": {
		detect: func(msg string) bool {
			return strings.Contains(msg, "Next message is the coin name")
		},
		parse: func(msg string) string {
			return strings.ToUpper(msg)
		},
	},
}

func (c *Chat) WaitForPumpCurrency(ctx context.Context, channel types.ChannelMessageIn) (string, error) {
	var (
		chatID     string
		parserName string
	)

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

			if parserName == "" {
				for chat, parser := range symbolParser {
					if parser.detect(msg.Text) {
						if msg.ChatID != nil {
							chatID = *msg.ChatID
						}
						parserName = chat
						fmt.Printf("PRE chat message for %s in chat id %s: %+v\n", parserName, chatID, msg)
						break
					}
				}
			} else if msg.ChatID != nil && chatID == *msg.ChatID && parserName != "" {
				fmt.Printf("Found after PREV message!\n")
				return symbolParser[parserName].parse(msg.Text), nil
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
