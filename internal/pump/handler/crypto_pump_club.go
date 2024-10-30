package handler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	chatTypes "trade_bot/internal/chat/types"
	clientTypes "trade_bot/internal/client/types"
	"trade_bot/internal/pump/types"
	commonTypes "trade_bot/internal/types"
)

const (
	cryptoPumpClubChatID string = "2321027763" // fake pump channel
	//cryptoPumpClubChatID string  = "1625691880" // real pump channel
	amountToSpend float64 = 2
)

type exchangeClient interface {
	GetPrice(ctx context.Context, symbol, baseSymbol string) (float64, error)
	CreateSpotOrder(ctx context.Context, order *clientTypes.SpotOrder) error
	GetAssets(ctx context.Context, symbol string) (float64, error)
	CancelAllOrders(ctx context.Context, symbol, baseSymbol string) error
}

type CryptoPumpClub struct {
	exchangeClient exchangeClient
	log            *logrus.Logger

	pump *types.Pump
}

func NewCryptoPumpClub(
	exchangeClient exchangeClient,
	log *logrus.Logger,
) *CryptoPumpClub {
	return &CryptoPumpClub{
		exchangeClient: exchangeClient,
		log:            log,
	}
}

func (c *CryptoPumpClub) CanHandle(ctx context.Context, msg *chatTypes.ChatIncomingMessage) bool {
	return msg.ChatID == cryptoPumpClubChatID
}

func (c *CryptoPumpClub) Process(ctx context.Context, msg *chatTypes.ChatIncomingMessage) error {
	switch {
	case c.pump == nil:
		// start point: receive "next is symbol" message
		if err := c.waitForNextMessage(ctx, msg.Text); err != nil {
			return fmt.Errorf("CryptoPumpClub::Process : %w", err)
		}
	case c.pump.Status == types.PumpStatusNextSymbol:
		// get pump symbol and process pump
		if err := c.waitForSymbol(ctx, msg.Text); err != nil {
			return fmt.Errorf("CryptoPumpClub::Process : %w", err)
		}

		if err := c.doPump(ctx, *c.pump); err != nil {
			return fmt.Errorf("CryptoPumpClub::Process : %w", err)
		}
	}

	return nil
}

func (c *CryptoPumpClub) waitForNextMessage(_ context.Context, msg string) error {
	isNextMessage := strings.Contains(msg, "Next message is the coin name")
	if !isNextMessage {
		return nil
	}

	c.pump = types.NewPump()
	c.pump.Status = types.PumpStatusNextSymbol

	c.log.Info("Next is coin message was received")

	return nil
}

func (c *CryptoPumpClub) waitForSymbol(_ context.Context, msg string) error {
	if c.pump == nil {
		return nil
	}

	symbol := strings.ToUpper(msg)
	baseSymbol := "USDT"

	c.pump.Symbol = &symbol
	c.pump.Status = types.PumpStatusSymbolReceived
	c.pump.BaseSymbol = &baseSymbol

	c.log.
		WithField("symbol", symbol).
		Info("Symbol was received")

	return nil
}

func (c *CryptoPumpClub) doPump(ctx context.Context, pump types.Pump) error {
	var wg sync.WaitGroup
	koef := []float64{1.05, 1.15, 1.25}
	n := len(koef)

	start := time.Now()
	startPrice, err := c.exchangeClient.GetPrice(ctx, *pump.Symbol, *pump.BaseSymbol)
	if err != nil {
		return fmt.Errorf("CryptoPumpClub::doPump : %w", err)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(coefficient float64) {
			defer wg.Done()
			ctxBuy, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			c.buyStock(ctxBuy, pump, coefficient)

			if ctxBuy.Err() == context.DeadlineExceeded {
				c.log.Error("CryptoPumpClub::doPump is stopped by timeout")
			} else {
				c.log.Info("Buy order created")
			}
		}(koef[i])
	}

	wg.Wait()

	c.log.WithField("timestamp", start.String()).Info("Wait some time")

	for {
		current := time.Now()
		if start.Minute() != current.Minute() {
			c.log.
				WithFields(logrus.Fields{
					"start_time":   start.String(),
					"current_time": current.String(),
				}).
				Info("Minutes are not equal")

			break
		} else if start.Minute() == current.Minute() && current.Second() > 34 {
			c.log.
				WithFields(logrus.Fields{
					"start_time":   start.String(),
					"current_time": current.String(),
				}).
				Info("Too many seconds passed")

			break
		}

		currentPrice, err := c.exchangeClient.GetPrice(ctx, *pump.Symbol, *pump.BaseSymbol)
		if err != nil {
			return fmt.Errorf("CryptoPumpClub::doPump : %w", err)
		}

		if startPrice*2 < currentPrice {
			c.log.
				WithFields(logrus.Fields{
					"start_price":        startPrice,
					"double_start_price": startPrice * 2,
					"current_price":      currentPrice,
				}).
				Info("Current price is double greater that start")

			break
		}

		c.log.
			WithFields(logrus.Fields{
				"start_price":   startPrice,
				"current_price": currentPrice,
				"current_time":  current.String(),
			}).
			Info("Price received")
	}

	c.exitPump(ctx, pump)

	return nil
}

func (c *CryptoPumpClub) buyStock(ctx context.Context, pump types.Pump, priceKoef float64) error {
	price, err := c.exchangeClient.GetPrice(ctx, *pump.Symbol, *pump.BaseSymbol)
	if err != nil {
		return fmt.Errorf("CryptoPumpClub::buyStock : %w", err)
	}

	price = price * priceKoef

	buyOrder := clientTypes.SpotOrder{
		Symbol:     *pump.Symbol,
		BaseSymbol: *pump.BaseSymbol,
		Position:   commonTypes.PositionLong,
		Type:       commonTypes.OrderTypeLimit,
		Entry:      price,
		Quantity:   amountToSpend / price,
	}
	if err = c.exchangeClient.CreateSpotOrder(ctx, &buyOrder); err != nil {
		return fmt.Errorf("CryptoPumpClub::buyStock : %w", err)
	}

	c.log.
		WithFields(logrus.Fields{
			"price":    buyOrder.Entry,
			"currency": buyOrder.Symbol + buyOrder.BaseSymbol,
			"quantity": buyOrder.Quantity,
			"type":     buyOrder.Type,
		}).
		Info("Buy order created")

	return nil
}

func (c *CryptoPumpClub) exitPump(ctx context.Context, pump types.Pump) error {
	if err := c.exchangeClient.CancelAllOrders(ctx, *pump.Symbol, *pump.BaseSymbol); err != nil {
		return fmt.Errorf("CryptoPumpClub::exitPump : %w", err)
	}

	c.log.Info("All orders cancelled")

	price, err := c.exchangeClient.GetPrice(ctx, *pump.Symbol, *pump.BaseSymbol)
	if err != nil {
		return fmt.Errorf("CryptoPumpClub::buyStock : %w", err)
	}

	c.log.Info("Got last price")

	qty, err := c.exchangeClient.GetAssets(ctx, *pump.Symbol)
	if err != nil {
		return fmt.Errorf("CryptoPumpClub::buyStock : %w", err)
	}

	c.log.Info("Got last asset quantity")

	order := clientTypes.SpotOrder{
		Symbol:     *pump.Symbol,
		BaseSymbol: *pump.BaseSymbol,
		Position:   commonTypes.PositionShort,
		Type:       commonTypes.OrderTypeMarket,
		Entry:      price * 0.95,
		Quantity:   qty,
	}
	if err = c.exchangeClient.CreateSpotOrder(ctx, &order); err != nil {
		return fmt.Errorf("CryptoPumpClub::buyStock : %w", err)
	}

	c.log.
		WithFields(logrus.Fields{
			"price":    order.Entry,
			"currency": order.Symbol + order.BaseSymbol,
			"quantity": order.Quantity,
			"type":     order.Type,
		}).
		Info("Sell order created")

	return nil
}
