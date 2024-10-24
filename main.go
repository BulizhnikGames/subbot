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

	tgBot, err := bot.StartBot(cfg, 10*time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = tgBot.Scraper.Run(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err = tgBot.SendNewPostsFromChannelToGroups(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	botErr := tgBot.WaitForUpdate(context.Background())
	if botErr != nil {
		log.Fatal(botErr)
	}
}
