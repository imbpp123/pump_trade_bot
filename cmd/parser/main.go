package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/sirupsen/logrus"

	"trade_bot/internal/chat/client"
	"trade_bot/internal/signals"
	"trade_bot/internal/signals/handler"
)

const (
	chatMessageTopic   string = "chat.income"
	signalMessageTopic string = "signal.created"
)

func main() {
	// Set up logger
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture termination signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Info("Received termination signal, shutting down gracefully...")
		cancel()
	}()

	// Initialize PubSub
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(true, false),
	)

	// Load environment variables
	appID, err := strconv.Atoi(os.Getenv("TELEGRAM_APP_ID"))
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_APP_ID: %v", err)
	}
	apiHash := os.Getenv("TELEGRAM_API_HASH")
	phone := os.Getenv("TELEGRAM_PHONE")
	sqliteDb := os.Getenv("TELEGRAM_SQLITE_DB")
	if apiHash == "" || phone == "" || sqliteDb == "" {
		log.Fatal("Environment variables TELEGRAM_API_HASH, TELEGRAM_PHONE, and TELEGRAM_SQLITE_DB must be set")
	}

	// Initialize Telegram client
	tgClient, err := client.NewTelegram(&client.TelegramOptions{
		AppID:        appID,
		ApiHash:      apiHash,
		Phone:        phone,
		SQLiteDb:     sqliteDb,
		Context:      ctx,
		Log:          log,
		Publisher:    pubSub,
		MessageTopic: chatMessageTopic,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Telegram client: %w", err)
	}

	// Start receiving messages
	if err := tgClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start Telegram client: %w", err)
	}

	log.Info("Waiting for messages...")

	// initialize message parser
	parser := signals.NewParser(&signals.ParserOptions{
		Handlers: []signals.Handler{
			handler.NewHardcoreVIP(),
		},
		Logger:            log,
		MessageSubscriber: pubSub,
		MessageTopic:      chatMessageTopic,
		SignalTopic:       signalMessageTopic,
		SignalPublisher:   pubSub,
	})
	if err := parser.Start(ctx); err != nil {
		log.Fatalf("Failed to start parsing messages: %w", err)
	}
}
