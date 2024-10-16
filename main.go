package main

import (
	"context"
	"github.com/BulizhnikGames/subbot/internal/bot"
	"github.com/BulizhnikGames/subbot/internal/config"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func main() {
	config.Load()
	cfg := config.Get()

	scraper, err := bot.StartScraper(cfg.API_ID, cfg.API_hash)
	if err != nil {
		log.Fatal(err)
	}

	tgBot, err := bot.StartBot(cfg.Bot_token, cfg.DB_URL, 10*time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		botErr := tgBot.WaitForUpdate(context.Background())
		if botErr != nil {
			log.Fatal(botErr)
		}
	}()

	err = scraper.Run(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
