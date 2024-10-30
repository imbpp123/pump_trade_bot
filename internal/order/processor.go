package order

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"

	signalTypes "trade_bot/internal/signals/types"
)

type orderHandler interface {
	ProcessSignal(ctx context.Context, signal *signalTypes.Signal) error
}

type ProcessorOptions struct {
	SignalTopic      string
	SignalSubscriber message.Subscriber
	Logger           *logrus.Logger
}

type Processor struct {
	signalTopic      string
	signalSubscriber message.Subscriber
	orderHandler     orderHandler
	log              *logrus.Logger
}

func NewProcessor(opt *ProcessorOptions) *Processor {
	return &Processor{
		signalTopic:      opt.SignalTopic,
		signalSubscriber: opt.SignalSubscriber,
		log:              opt.Logger,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	messages, err := p.signalSubscriber.Subscribe(ctx, p.signalTopic)
	if err != nil {
		p.log.
			WithError(err).
			Error("Failed to subscribe to topic")

		return fmt.Errorf("Processor::Start : %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			p.log.Info("Processor context cancelled, stopping order processing")
			return nil
		case rawMsg, ok := <-messages:
			if !ok {
				p.log.Warn("Message channel closed, stopping message parsing")
				return nil
			}

			log := p.log.WithFields(logrus.Fields{
				"MessageUUID": rawMsg.UUID,
			})

			var msg signalTypes.Signal
			if err := json.Unmarshal(rawMsg.Payload, &msg); err != nil {
				log.
					WithError(err).
					Error("Failed to unmarshal incoming message")

				continue
			}

			log = log.WithFields(logrus.Fields{
				"Channel":  msg.Channel,
				"Exchange": msg.Exchange,
				"Position": msg.Position,
				"Symbol":   msg.Symbol,
			})
			log.Debug("Processing incoming signal")

			// process signal to order
			if err := p.orderHandler.ProcessSignal(ctx, &msg); err != nil {
				log.WithError(err).Error("Failed to process signal into order")

				continue
			}

			rawMsg.Ack()
		}
	}
}

func (p *Processor) Stop(ctx context.Context) error {
	p.log.Info("Stopping order processor")

	return nil
}
