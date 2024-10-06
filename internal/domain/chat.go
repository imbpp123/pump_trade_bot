package domain

import (
	"context"
	"fmt"
	"log"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
)

type chatClient interface {
}

type Chat struct {
	client chatClient
}

func NewChat(client chatClient) *Chat {
	return &Chat{
		client: client,
	}
}

func (c *Chat) WaitForPumpMessage(ctx context.Context, channel string) error {
	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		4723758,
		// Get ApiHash from https://my.telegram.org/apps
		"0fb8a49cb3655abfbb77d311627d106f",
		// ClientType, as we defined above
		gotgproto.ClientTypePhone("+4915257015190"),
		// Optional parameters of client
		&gotgproto.ClientOpts{
			//Session: sessionMaker.SimpleSession(),
			Session: sessionMaker.SqlSession(sqlite.Open("userbot")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	clientDispatcher := client.Dispatcher
	clientDispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, download), 1)

	err = client.Idle()
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	return nil
}

func download(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage

	fmt.Println(msg.FromID, msg.PeerID, msg.Text)
	return nil
}
