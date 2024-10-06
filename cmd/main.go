package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gofor-little/env"

	"trade_bot/internal/client"
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

	mexcClient := client.NewMexc(
		env.Get("MEXC_API_KEY", ""),
		env.Get("MEXC_API_SECRET", ""),
	)
	fmt.Println(mexcClient.GetCurrencyPriceTicker(ctx, "BTCUSDT"))

	/*
		currency := "BTCUSDT"
		candleRepository := repository.NewCandle()

		candleService := domain.NewCandle(mexcClient, candleRepository)

		if err := candleService.StartUpdate(ctx); err != nil {
			panic(err)
		}
		candleService.SubscribeForUpdate(ctx, currency)
		time.Sleep(time.Second)
		fmt.Println(candleService.GetCandle(ctx, currency))
	*/
	/*
		// run account asset update for repository in background
		assetService := domain.NewAsset(mexcClient, repository.NewAsset())
		go func() {
			if err := assetService.RunUpdate(ctx); err != nil {
				panic(err)
			}
		}()

		// run telegram as chat
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

		// let's do the pump!

		// get currency from chat
		chatService := domain.NewChat()
		currency, err := chatService.WaitForPumpCurrency(ctx, chanMessageIn)
		if err != nil {
			panic(err)
		}

		// do the job!


		// stop everything
		tgClient.Stop(ctx)
		close(chanMessageIn)

	*/
}
