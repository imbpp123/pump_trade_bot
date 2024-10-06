package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"

	"trade_bot/internal/types"
)

type TelegramOptions struct {
	AppID    int
	ApiHash  string
	Phone    string
	SQLiteDb string
	Context  context.Context
}

type Telegram struct {
	client    *gotgproto.Client
	isStarted bool
}

var ErrTelegramStarted = errors.New("already started")

func NewTelegram(opt TelegramOptions) (*Telegram, error) {
	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		opt.AppID,
		// Get ApiHash from https://my.telegram.org/apps
		opt.ApiHash,
		gotgproto.ClientTypePhone(opt.Phone),
		&gotgproto.ClientOpts{
			//Session: sessionMaker.SimpleSession(), // in memory
			Session: sessionMaker.SqlSession(sqlite.Open(opt.SQLiteDb)),
			Context: opt.Context,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("NewTelegram : %w", err)
	}

	return &Telegram{
		client:    client,
		isStarted: false,
	}, nil
}

func (t *Telegram) StartRecvMessages(_ context.Context, msgChan types.ChannelMessageIn) error {
	if t.isStarted {
		return ErrTelegramStarted
	}
	t.isStarted = true

	fmt.Println("Started to receive messages")

	t.client.Dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, func(ctx *ext.Context, update *ext.Update) error {
		msg := update.EffectiveMessage

		message := types.ChatMessageIn{
			Text: msg.Text,
		}

		chat := update.GetChannel()
		if chat != nil {
			str := strconv.FormatInt(chat.ID, 10)
			message.ChatID = &str
		}

		user := update.GetUserChat()
		if user != nil {
			str := strconv.FormatInt(user.ID, 10)
			message.UserID = &str
		}

		msgChan <- &message

		return nil
	}), 1)

	return nil
}

func (t *Telegram) Stop(_ context.Context) error {
	t.client.Stop()
	return nil
}
