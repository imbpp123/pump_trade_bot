package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofor-little/env"
	"github.com/sirupsen/logrus"

	"trade_bot/internal/chat/client"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := env.Load(dir + "/.env"); err != nil {
		panic(err)
	}

	appIDStr, err := strconv.Atoi(env.Get("TELEGRAM_APP_ID", ""))
	if err != nil {
		panic(err)
	}
	tgClient, err := client.NewTelegram(&client.TelegramOptions{
		AppID:        appIDStr,
		ApiHash:      env.Get("TELEGRAM_API_HASH", ""),
		Phone:        env.Get("TELEGRAM_PHONE", ""),
		SQLiteDb:     env.Get("TELEGRAM_SQLITE_DB", ""),
		Context:      ctx,
		Log:          log,
		Publisher:    pubSub,
		MessageTopic: "chat.income",
	})
	if err != nil {
		panic(err)
	}
	tgClient.Start(ctx)

	fmt.Println("Wait for messages")

	messages, err := pubSub.Subscribe(context.Background(), "chat.income")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		fmt.Printf("received message: %s, payload: %s\n", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}
