package signals

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"

	chatTypes "trade_bot/internal/chat/types"
	"trade_bot/internal/signals/types"
)

type Handler interface {
	Name() string
	CanHandle(ctx context.Context, message *chatTypes.ChatIncomingMessage) bool
	ParseSignal(ctx context.Context, message *chatTypes.ChatIncomingMessage) (*types.Signal, error)
}

type ParserOptions struct {
	Handlers          []Handler
	Logger            *logrus.Logger
	MessageSubscriber message.Subscriber
	MessageTopic      string
	SignalTopic       string
	SignalPublisher   message.Publisher
}

type Parser struct {
	handlers          []Handler
	log               *logrus.Logger
	messageTopic      string
	messageSubscriber message.Subscriber
	signalTopic       string
	signalPublisher   message.Publisher

	messages <-chan *message.Message
}

func NewParser(opt *ParserOptions) *Parser {
	return &Parser{
		handlers:          opt.Handlers,
		log:               opt.Logger,
		messageTopic:      opt.MessageTopic,
		messageSubscriber: opt.MessageSubscriber,
		signalTopic:       opt.SignalTopic,
		signalPublisher:   opt.SignalPublisher,
	}
}

func (p *Parser) Start(ctx context.Context) error {
	var err error

	p.messages, err = p.messageSubscriber.Subscribe(ctx, p.messageTopic)
	if err != nil {
		p.log.
			WithError(err).
			Error("Failed to subscribe to topic")

		return fmt.Errorf("Parser::Parse : %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			p.log.Info("Parser context cancelled, stopping message parsing")
			return nil
		case rawMsg, ok := <-p.messages:
			if !ok {
				p.log.Warn("Message channel closed, stopping message parsing")
				return nil
			}

			log := p.log.WithFields(logrus.Fields{
				"UUID": rawMsg.UUID,
			})

			var msg chatTypes.ChatIncomingMessage
			if err := json.Unmarshal(rawMsg.Payload, &msg); err != nil {
				log.
					WithError(err).
					Error("Failed to unmarshal incoming message")

				continue
			}

			log = log.WithFields(logrus.Fields{
				"ChatID": msg.ChatID,
				"UserID": msg.UserID,
				"Text":   msg.Text,
			})
			log.Debug("Parsing incoming message")

			var handler Handler
			for _, h := range p.handlers {
				if h.CanHandle(ctx, &msg) {
					handler = h
					break
				}
			}

			if handler == nil {
				log.Debug("No handler found for message")

				rawMsg.Ack()
				continue
			}

			log = log.WithFields(logrus.Fields{
				"Handler": handler.Name(),
			})

			signal, err := handler.ParseSignal(ctx, &msg)
			if err != nil {
				log.
					WithError(err).
					Error("Failed to parse signal")

				rawMsg.Ack()
				continue
			}

			log.Info("Message parsed to signal successfully")

			// send signal to topic
			rawMessage, err := json.Marshal(signal)
			if err != nil {
				log.
					WithError(err).
					Error("Failed to marshal signal to JSON")

				return fmt.Errorf("Parser::Start : %w", err)
			}

			if err := p.signalPublisher.Publish(p.signalTopic, message.NewMessage(watermill.NewUUID(), rawMessage)); err != nil {
				log.
					WithError(err).
					Error("Failed to publish message to topic")

				return fmt.Errorf("Parser::Start : %w", err)
			}

			// Acknowledge the message to mark it as processed
			rawMsg.Ack()
		}
	}
}

func (p *Parser) Stop(ctx context.Context) error {
	p.log.Info("Stopping parser")

	return nil
}
