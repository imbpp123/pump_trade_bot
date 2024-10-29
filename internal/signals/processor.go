package signals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"

	chatTypes "trade_bot/internal/chat/types"
	"trade_bot/internal/signals/types"
	commonTypes "trade_bot/internal/types"
)

type signalRepository interface {
	Create(ctx context.Context, signal *types.Signal) error
}

type Handler interface {
	Name() commonTypes.SignalChannel
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
	SignalRepository  signalRepository
}

type Parser struct {
	handlers          []Handler
	log               *logrus.Logger
	messageTopic      string
	messageSubscriber message.Subscriber
	signalTopic       string
	signalPublisher   message.Publisher
	signalRepository  signalRepository
}

func NewParser(opt *ParserOptions) *Parser {
	return &Parser{
		handlers:          opt.Handlers,
		log:               opt.Logger,
		messageTopic:      opt.MessageTopic,
		messageSubscriber: opt.MessageSubscriber,
		signalTopic:       opt.SignalTopic,
		signalPublisher:   opt.SignalPublisher,
		signalRepository:  opt.SignalRepository,
	}
}

func (p *Parser) Start(ctx context.Context) error {
	messages, err := p.messageSubscriber.Subscribe(ctx, p.messageTopic)
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
			})

			// handle message
			err := p.processSignal(ctx, &msg, log)
			if err != nil {
				switch {
				case errors.Is(err, types.ErrSignalHandlerNotFound):
					log.WithError(err).Debug("Handler is not found")
				default:
					log.WithError(err).Error("Failed to handle signal message")
				}
			}

			rawMsg.Ack()
		}
	}
}

func (p *Parser) Stop(ctx context.Context) error {
	p.log.Info("Stopping parser")

	return nil
}

func (p *Parser) processSignal(ctx context.Context, signalMessage *chatTypes.ChatIncomingMessage, log *logrus.Entry) error {
	// find handler
	handler, err := p.findHandler(ctx, signalMessage)
	if err != nil {
		return fmt.Errorf("Parser::processSignal : %w", err)
	}
	log = log.WithFields(logrus.Fields{
		"MessageHandler": handler.Name(),
	})
	log.Debug("Handler parser is found")

	// parse text to signal
	signal, err := handler.ParseSignal(ctx, signalMessage)
	if err != nil {
		return fmt.Errorf("Parser::processSignal : %w", err)
	}
	log.Debug("Message parsed to signal")

	// save signal to storage
	if err := p.signalRepository.Create(ctx, signal); err != nil {
		return fmt.Errorf("Parser::processSignal : %w", err)
	}
	log.Debug("Signal saved to storage")

	// send signal to topic
	rawMessage, err := json.Marshal(signal)
	if err != nil {
		return fmt.Errorf("Parser::processSignal : %w", err)
	}
	if err := p.signalPublisher.Publish(p.signalTopic, message.NewMessage(watermill.NewUUID(), rawMessage)); err != nil {
		return fmt.Errorf("Parser::processSignal : %w", err)
	}
	log.
		WithField("signal", rawMessage).
		Debug("Message signal is processed")

	return nil
}

func (p *Parser) findHandler(ctx context.Context, signalMessage *chatTypes.ChatIncomingMessage) (Handler, error) {
	var handler Handler

	for _, h := range p.handlers {
		if h.CanHandle(ctx, signalMessage) {
			handler = h
			break
		}
	}
	if handler == nil {
		return nil, types.ErrSignalHandlerNotFound
	}

	return handler, nil
}
