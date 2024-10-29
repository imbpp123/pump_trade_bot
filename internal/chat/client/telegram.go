package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"

	"trade_bot/internal/chat/types"
)

const handlerGroupMessage int = 0

// TelegramOptions holds configuration options for the Telegram client.
type TelegramOptions struct {
	AppID        int
	ApiHash      string
	Phone        string
	SQLiteDb     string
	Context      context.Context
	Publisher    message.Publisher
	Log          *logrus.Logger
	MessageTopic string
}

// Telegram represents a client for interacting with Telegram messages.
type Telegram struct {
	client           *gotgproto.Client
	isStarted        bool
	log              *logrus.Logger
	messagePublisher message.Publisher
	messageTopic     string
}

// NewTelegram creates a new Telegram client with the provided options.
// If a logger is not specified in TelegramOptions, it initializes a new logrus instance.
func NewTelegram(opt *TelegramOptions) (*Telegram, error) {
	log := opt.Log
	if log == nil {
		log = logrus.New()
		log.SetLevel(logrus.InfoLevel)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		opt.AppID,
		// Get ApiHash from https://my.telegram.org/apps
		opt.ApiHash,
		gotgproto.ClientTypePhone(opt.Phone),
		&gotgproto.ClientOpts{
			Session: sessionMaker.SqlSession(sqlite.Open(opt.SQLiteDb)),
			Context: opt.Context,
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to create Telegram client")

		return nil, fmt.Errorf("NewTelegram : %w", err)
	}

	log.WithFields(logrus.Fields{
		"AppID":    opt.AppID,
		"SQLiteDb": opt.SQLiteDb,
	}).Info("Telegram client created successfully")

	return &Telegram{
		client:           client,
		isStarted:        false,
		log:              log,
		messagePublisher: opt.Publisher,
		messageTopic:     opt.MessageTopic,
	}, nil
}

// Start initializes message reception on the Telegram client.
// It takes a channel to send incoming messages and returns an types.ErrChatStarted error if the client is already started.
func (t *Telegram) Start(_ context.Context) error {
	if t.isStarted {
		t.log.Warn("Attempted to start receiving messages, but client is already started")

		return types.ErrChatStarted
	}
	t.isStarted = true

	t.log.Info("Started to receive messages")

	t.client.Dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, t.messageHandler), handlerGroupMessage)

	return nil
}

// Stop stops the Telegram client
func (t *Telegram) Stop(_ context.Context) {
	if !t.isStarted {
		t.log.Warn("Attempted to stop Telegram client, but it is not started")
		return
	}

	t.client.Stop()
	t.isStarted = false
	t.log.Info("Stopped receiving messages")
}

func (t *Telegram) messageHandler(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage
	if msg == nil {
		t.log.
			WithError(types.ErrChatNoMessage).
			Error("Received update with no message")

		return nil
	}

	chatID := ""
	if chat := update.GetChannel(); chat != nil {
		chatID = strconv.FormatInt(chat.ID, 10)
	}

	chatMessage := types.NewChatIncomingMessage(
		types.ChatTypeTelegram,
		msg.Text,
		chatID,
	)

	rawMessage, err := json.Marshal(chatMessage)
	if err != nil {
		t.log.
			WithError(err).
			Error("Failed to marshal chat message to JSON")

		return fmt.Errorf("Telegram::messageHandler : %w", err)
	}

	if err := t.messagePublisher.Publish(t.messageTopic, message.NewMessage(watermill.NewUUID(), rawMessage)); err != nil {
		t.log.
			WithError(err).
			Error("Failed to publish message")

		return fmt.Errorf("Telegram::messageHandler : %w", err)
	}

	return nil
}
