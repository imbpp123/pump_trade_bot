package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/gofor-little/env"

	"trade_bot/internal/client"
	"trade_bot/internal/domain"
	"trade_bot/internal/types"
)

const (
	amountToSpend  float64 = 1.2
	orderBuyKoeff  float64 = 1.05
	orderSellKoeff float64 = 0.95

	downForSellPercent int     = 100
	profitPercent      float64 = 1
)

type exchangeClient interface {
	CreateOrder(ctx context.Context, order *types.OrderCreate) (types.OrderID, error)
	CreateBatchOrders(ctx context.Context, currency string, side types.OrderSide, orderType types.OrderType, orders []types.OrderCreate) error
	CancelAllOrders(ctx context.Context, currency string) error
	GetAssets(ctx context.Context, currency string) (float64, error)
	GetCurrencyCandles(ctx context.Context, currency string, interval types.CandleInterval) ([]types.Candle, error)
	GetCurrencyPriceTicker(ctx context.Context, currency string) (float64, error)
}

var (
	mexcClient exchangeClient
)

func main() {
	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := env.Load(dir + "/../.env"); err != nil {
		panic(err)
	}

	// mexc client
	mexcClient = client.NewMexc(
		env.Get("MEXC_API_KEY", ""),
		env.Get("MEXC_API_SECRET", ""),
	)

	// telegram wait for pump currency
	currency, timestamp := getPumpCurrency(ctx)
	currencyUSDT := currency + "USDT"
	fmt.Printf("crypto is set to %s at %s\n", currencyUSDT, timestamp.String())

	// create limit order for given currency
	priceBuy := getPrice(ctx, currencyUSDT) * orderBuyKoeff
	buyOrder := types.OrderCreate{
		Currency: currencyUSDT,
		Side:     types.OrderSideLong,
		Type:     types.OrderTypeLimit,
		Price:    priceBuy,
		Quantity: amountToSpend / priceBuy,
	}
	fmt.Printf("Buy order %+v total = %f\n", buyOrder, buyOrder.Quantity*buyOrder.Price)
	_, err = mexcClient.CreateOrder(ctx, &buyOrder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("order was created\n")

	// wait for order to process
	var boughtQTY float64
	for {
		boughtQTY, err = mexcClient.GetAssets(ctx, currency)
		if err != nil && !errors.Is(err, client.ErrAssetNotFound) {
			panic(err)
		}
		if math.Ceil(boughtQTY) >= math.Ceil(buyOrder.Quantity) {
			// bought!
			fmt.Printf("we bought %f amount of %s\n", boughtQTY, currency)
			break
		}
		// we can make 2 requests for MEXC in 1 second - this logic should not be here!!!
		time.Sleep(time.Millisecond * 400)
	}

	// create 3 SHORT LIMIT orders
	orderAmount := 3
	shortOrders := make([]types.OrderCreate, orderAmount)
	shortQTY := float64(0)
	fmt.Printf("Create %d SHORT LIMIT orders:\n", orderAmount)
	for i := 0; i < orderAmount; i++ {
		shortOrders[i].Currency = currencyUSDT
		shortOrders[i].Price = buyOrder.Price + buyOrder.Price*(profitPercent+float64(i)*0.1)
		shortOrders[i].Quantity = math.Ceil((boughtQTY/float64(orderAmount))*100) / 100
		shortOrders[i].Side = types.OrderSideShort
		shortOrders[i].Type = types.OrderTypeLimit

		if i < orderAmount-1 {
			shortQTY += shortOrders[i].Quantity
		} else {
			shortOrders[i].Quantity = math.Ceil((boughtQTY-shortQTY)*100) / 100
		}
		fmt.Printf("%+v\n", shortOrders[i])
	}
	if err := mexcClient.CreateBatchOrders(ctx, currencyUSDT, types.OrderSideShort, types.OrderTypeLimit, shortOrders); err != nil {
		panic(err)
	}

	// close order and sell everything even if we fail!
	defer func() {
		fmt.Println("SELL!")

		// sell and close all orders if any
		// cancel orders.. no need to buy anymore
		mexcClient.CancelAllOrders(ctx, currencyUSDT)

		// sell all what we possibly have
		sellOrder := types.OrderCreate{
			Currency: currencyUSDT,
			Side:     types.OrderSideShort,
			Type:     types.OrderTypeLimit,
			Price:    getPrice(ctx, currencyUSDT) * orderSellKoeff,
			Quantity: getAssetQTY(ctx, currency),
		}

		if sellOrder.Quantity == 0 {
			fmt.Println("Nothing to sell!")
			return
		}

		fmt.Printf("Sell order %+v total = %f\n", sellOrder, sellOrder.Quantity*sellOrder.Price)
		_, err = mexcClient.CreateOrder(ctx, &sellOrder)
		if err != nil {
			panic(err)
		}
		fmt.Printf("order was created\n")
		fmt.Println("DONE!")
	}()

	// wait for a moment to sell
	for {
		start := time.Now()
		fmt.Printf("Current time: %s\n", start.String())

		if timestamp.Minute() != start.Minute() {
			// new minute - SELL!
			break
		} else if timestamp.Minute() == start.Minute() && start.Second() > 42 {
			// 10 seconds left in this minute - SELL!!
			break
		}

		candles, err := mexcClient.GetCurrencyCandles(ctx, currencyUSDT, types.CandleInterval1m)
		if err != nil {
			panic(err)
		}

		var currCandle *types.Candle

		for _, candle := range candles {
			if candle.OpenTime.Before(timestamp) && candle.CloseTime.After(timestamp) {
				currCandle = &candle
			}
		}
		if currCandle == nil {
			// no candle - SELL!
			break
		}

		currPrice := getPrice(ctx, currencyUSDT)

		fmt.Printf(
			"curr = %s candle = %+v, price = %f, time = %s\n",
			time.Now().String(),
			currCandle,
			priceBuy,
			time.Since(start).String(),
		)

		if currPrice/currCandle.High < float64(downForSellPercent/100) {
			// we have decrease in price - SELL!
			break
		}
	}
	// ---- WAIT FOR A MOMENT TO SELL ENDED! SEEEEELLLL!!!
}

func getPumpCurrency(ctx context.Context) (string, time.Time) {
	chanMessageIn := make(types.ChannelMessageIn)
	appIDStr, err := strconv.Atoi(env.Get("TELEGRAM_APP_ID", ""))
	if err != nil {
		panic(err)
	}
	tgClient, err := client.NewTelegram(client.TelegramOptions{
		AppID:    appIDStr,
		ApiHash:  env.Get("TELEGRAM_API_HASH", ""),
		Phone:    env.Get("TELEGRAM_PHONE", ""),
		SQLiteDb: env.Get("TELEGRAM_SQLITE_DB", ""),
		Context:  ctx,
	})
	if err != nil {
		panic(err)
	}
	tgClient.StartRecvMessages(ctx, chanMessageIn)

	// get currency from chat
	chatService := domain.NewChat()
	currency, err := chatService.WaitForPumpCurrency(ctx, chanMessageIn)
	if err != nil {
		panic(err)
	}

	tgClient.Stop(ctx)
	close(chanMessageIn)

	return currency, time.Now()
}

func getPrice(ctx context.Context, currency string) float64 {
	price, err := mexcClient.GetCurrencyPriceTicker(ctx, currency)
	if err != nil {
		panic(err)
	}

	return price
}

func getAssetQTY(ctx context.Context, currency string) float64 {
	qty, err := mexcClient.GetAssets(ctx, currency)
	if err != nil {
		panic(err)
	}

	return qty
}
