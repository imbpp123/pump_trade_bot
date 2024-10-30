package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"trade_bot/internal/chat/client"
	"trade_bot/internal/signals"
	"trade_bot/internal/signals/parser"
	"trade_bot/internal/signals/repository"
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

	// initialize gorm storage
	db, err := gorm.Open(sqlite.Open(os.Getenv("SQLITE_DATABASE")), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

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
		log.Fatalf("Failed to initialize Telegram client: %v", err)
	}

	// Start receiving messages
	if err := tgClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start Telegram client: %v", err)
	}

	log.Info("Waiting for messages...")

	// initialize message parser
	signalRepository, err := repository.NewGormSignal(db)
	if err != nil {
		log.Fatalf("Failed to create signal repository: %v", err)
	}
	parser := signals.NewParser(&signals.ParserOptions{
		Handlers: []signals.Handler{
			parser.NewHardcoreVIP(),
		},
		Logger:            log,
		MessageSubscriber: pubSub,
		MessageTopic:      chatMessageTopic,
		SignalTopic:       signalMessageTopic,
		SignalPublisher:   pubSub,
		SignalRepository:  signalRepository,
	})
	if err := parser.Start(ctx); err != nil {
		log.Fatalf("Failed to parse messages: %v", err)
	}
}

func StartTelegram(ctx context.Context, pubSub message.Publisher, log *logrus.Logger) {
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
		log.Fatalf("Failed to initialize Telegram client: %v", err)
	}

	// Start receiving messages
	if err := tgClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start Telegram client: %v", err)
	}

	log.Info("Waiting for messages...")
}
