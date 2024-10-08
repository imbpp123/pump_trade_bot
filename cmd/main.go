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
	AmountToSpend float64 = 1.1
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
	fmt.Printf("crypto is set to %s\n", currencyUSDT)

	// get price for currency
	price := getPrice(currencyUSDT)
	fmt.Printf("crypto price is %f\n", price)

	// hehe.. qty that we will pump
	qty := AmountToSpend / price
	fmt.Printf("QTY to buy is %f\n", qty)

	// create limit order for given currency
	buyOrder := types.OrderCreate{
		Currency: currencyUSDT,
		Side:     types.OrderSideLong,
		Type:     types.OrderTypeLimit,
		Price:    getPrice(currencyUSDT) * 0.95,
		Quantity: qty,
	}
	fmt.Printf("%+v\n", buyOrder)
	_, err = mexcClient.CreateOrder(ctx, &types.OrderCreate{
		Currency: currencyUSDT,
		Side:     types.OrderSideLong,
		Type:     types.OrderTypeLimit,
		Price:    price * 1.02,
		Quantity: qty,
	})
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
			Price:    getPrice(currencyUSDT) * 0.95,
			Quantity: getAssetQTY(currency),
		}
		fmt.Printf("%+v\n", sellOrder)
		_, err = mexcClient.CreateOrder(ctx, &sellOrder)
		if err != nil {
			panic(err)
		}
		fmt.Printf("order was created\n")
		fmt.Println("DONE!")
	}()

	// wait for a moment to sell
	for {
		candles, err := mexcClient.GetCurrencyCandles(ctx, currency, types.CandleInterval1m)
		if err != nil {
			panic(err)
		}

		for _, c := range candles {
			if c.OpenTime.Before(timestamp) && c.CloseTime.After(timestamp) {
				return c, nil
			}
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
