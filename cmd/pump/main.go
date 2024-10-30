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
	"github.com/sirupsen/logrus"

	"trade_bot/internal/chat/client"
	exchangeClient "trade_bot/internal/client"
	"trade_bot/internal/pump"
	"trade_bot/internal/pump/handler"
)

const (
	chatMessageTopic string = "chat.income"
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

	// start telegram
	StartTelegram(ctx, pubSub, log)

	// mexc client
	mexcClient := exchangeClient.NewMexc(
		os.Getenv("MEXC_API_KEY"),
		os.Getenv("MEXC_API_SECRET"),
	)

	processor := pump.NewProcessor(&pump.ProcessorOptions{
		MessageSubscriber: pubSub,
		MessageTopic:      chatMessageTopic,
		Log:               log,
		Handlers: []pump.PumpHandler{
			handler.NewCryptoPumpClub(mexcClient, log),
		},
	})
	if err := processor.Start(ctx); err != nil {
		log.Fatalf("Failed to process pump: %v", err)
	}
}

func StartTelegram(ctx context.Context, pubSub message.Publisher, log *logrus.Logger) *client.Telegram {
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

	return tgClient
}
