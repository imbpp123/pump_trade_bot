package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofor-little/env"

	"trade_bot/internal/client"
	"trade_bot/internal/domain"
	"trade_bot/internal/types"
)

const (
	amountToSpend  float64 = 1
	orderBuyKoeff  float64 = 1.2
	orderSellKoeff float64 = 0.8
)

var (
	mexcClient *client.Mexc
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
	priceBuy := getPrice(currencyUSDT) * orderBuyKoeff
	buyOrder := types.OrderCreate{
		Currency: currencyUSDT,
		Side:     types.OrderSideLong,
		Type:     types.OrderTypeLimit,
		Price:    priceBuy,
		Quantity: amountToSpend / priceBuy,
	}
	fmt.Printf("Buy order %+v\n", buyOrder)
	_, err = mexcClient.CreateOrder(ctx, &buyOrder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("order was created\n")

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
			Price:    getPrice(currencyUSDT) * orderSellKoeff,
			Quantity: getAssetQTY(currency),
		}
		fmt.Printf("Sell order %+v\n", sellOrder)
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
		} else if timestamp.Minute() == start.Minute() && start.Second() > 50 {
			// 10 seconds left in this minute - SELL!!
			break
		}

		candles, err := mexcClient.GetCurrencyCandles(currencyUSDT, types.CandleInterval1m)
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

		currPrice := getPrice(currencyUSDT)

		fmt.Printf(
			"curr = %s candle = %+v, price = %f, time = %s\n",
			time.Now().String(),
			currCandle,
			priceBuy,
			time.Since(start).String(),
		)

		if currPrice/priceBuy > 5 {
			// don't be greedy 500% is enough - SELL!
			break
		}
		if currPrice/currCandle.High < 0.7 {
			// we have 30% decrease in price - SELL!
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

func getPrice(currency string) float64 {
	price, err := mexcClient.GetCurrencyPriceTicker(currency)
	if err != nil {
		panic(err)
	}

	return price
}

func getAssetQTY(currency string) float64 {
	qty, err := mexcClient.GetAssets(currency)
	if err != nil {
		panic(err)
	}

	return qty
}
