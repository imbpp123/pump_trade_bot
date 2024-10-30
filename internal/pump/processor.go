package pump

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"

	chatTypes "trade_bot/internal/chat/types"
)

type PumpHandler interface {
	CanHandle(ctx context.Context, message *chatTypes.ChatIncomingMessage) bool
	Process(ctx context.Context, message *chatTypes.ChatIncomingMessage) error
}

type ProcessorOptions struct {
	MessageSubscriber message.Subscriber
	MessageTopic      string
	Log               *logrus.Logger
	Handlers          []PumpHandler
}

type Processor struct {
	messageSubscriber message.Subscriber
	messageTopic      string
	log               *logrus.Logger
	handlers          []PumpHandler
}

func NewProcessor(opt *ProcessorOptions) *Processor {
	return &Processor{
		messageSubscriber: opt.MessageSubscriber,
		messageTopic:      opt.MessageTopic,
		log:               opt.Log,
		handlers:          opt.Handlers,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	messages, err := p.messageSubscriber.Subscribe(ctx, p.messageTopic)
	if err != nil {
		p.log.
			WithError(err).
			Error("Failed to subscribe to topic")

		return fmt.Errorf("Processor::Start : %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			p.log.Info("Parser context cancelled, stopping message parsing")
			return nil
		case rawMsg, ok := <-messages:
			if !ok {
				p.log.Warn("Message channel closed, stopping message parsing")
				return nil
			}

			log := p.log.WithFields(logrus.Fields{
				"MessageUUID": rawMsg.UUID,
			})

			// parse text to incoming message
			var msg chatTypes.ChatIncomingMessage
			if err := json.Unmarshal(rawMsg.Payload, &msg); err != nil {
				log.
					WithError(err).
					Error("Failed to unmarshal incoming message")

				rawMsg.Ack()
				continue
			}
			log = log.WithFields(logrus.Fields{
				"IncomingMessageUUID": msg.UUID.String(),
				"ChatID":              msg.ChatID,
				"Text":                msg.Text,
				"CreatedAt":           msg.CreatedAt.String(),
			})

			if msg.CreatedAt.Before(time.Now().Add(time.Second * -10)) {
				log.Warn("Too late for pump")

				rawMsg.Ack()
				continue
			}

			// handle message
			err := p.processSignal(ctx, &msg, log)
			if err != nil {
				log.WithError(err).Error("Failed to handle signal message")
			}

			rawMsg.Ack()
		}
	}
}

func (p *Processor) Stop(ctx context.Context) error {
	p.log.Info("Stopping parser")

	return nil
}

func (p *Processor) processSignal(ctx context.Context, signalMessage *chatTypes.ChatIncomingMessage, log *logrus.Entry) error {
	for _, h := range p.handlers {
		if h.CanHandle(ctx, signalMessage) {
			if err := h.Process(ctx, signalMessage); err != nil {
				return fmt.Errorf("Processor::processSignal : %w", err)
			}

			break
		}
	}

	return nil
}
