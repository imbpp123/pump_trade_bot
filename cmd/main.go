package main

import (
	"context"
	"os"

	"github.com/gofor-little/env"

	"trade_bot/internal/domain"
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

	/*
		mexcClient := client.NewMexc(
			env.Get("MEXC_API_KEY", ""),
			env.Get("MEXC_API_SECRET", ""),
		)

		assetRepository := repository.NewAsset()
		assetService := domain.NewAsset(mexcClient, assetRepository)

		// run account asset update for repository in background
		go func() {
			if err := assetService.RunUpdate(ctx); err != nil {
				panic(err)
			}
		}()
	*/
	
	tgClient := domain.NewChat(nil)
	tgClient.WaitForPumpMessage(ctx, "test")

	/*
		// run tg channel check for crypto name in background
		tgClient := client.NewTelegram(
			env.Get("TELEGRAM_API_ID", ""),
			env.Get("TELEGRAM_API_HASH", ""),
		)

		tgService := domain.NewTelegram(tgClient)
		tgService.WaitForCurrency(ctx)

		time.Sleep(time.Minute)
	*/
	/*
		// run tg channel check for crypto name in background
		tg := client.NewTelegram()

		tgService := domain.NewTelegram(tg)
		tgService.WaitForCurrency(ctx)

		// run process pump-dump
		pumpService := domain.NewPump(tgService, assetService)
		pumpService.PumpDumpIt(ctx)
	*/
}
