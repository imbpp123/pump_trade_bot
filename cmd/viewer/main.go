package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gofor-little/env"

	"trade_bot/internal/client"
	"trade_bot/internal/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := env.Load(dir + "/.env"); err != nil {
		panic(err)
	}

	chanMessageIn := make(types.ChannelMessageIn)
	appIDStr, err := strconv.Atoi(env.Get("TELEGRAM_APP_ID", ""))
	if err != nil {
		panic(err)
	}
	tgClient, err := client.NewTelegram(client.TelegramOptions{
		AppID:    appIDStr,
		ApiHash:  env.Get("TELEGRAM_API_HASH", ""),
		Phone:    env.Get("TELEGRAM_PHONE", ""),
		SQLiteDb: env.Get("TELEGRAM_SQLITE_DB", ""),
		Context:  ctx,
	})
	if err != nil {
		panic(err)
	}
	tgClient.StartRecvMessages(ctx, chanMessageIn)

	fmt.Println("Wait for messages")

	for {
		select {
		case <-ctx.Done():
			close(chanMessageIn)

			fmt.Println(ctx.Err())
			return
		case msg := <-chanMessageIn:
			if msg == nil || msg.ChatID == nil {
				continue
			}

			fmt.Printf("chat message in chat_id = %s: %+v \n", *msg.ChatID, msg)
		}
	}
}
