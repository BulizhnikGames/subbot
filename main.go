package main

import (
	"github.com/BulizhnikGames/subbot/internal/config"
	"github.com/BulizhnikGames/subbot/internal/scraper"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	config.Load()
	cfg := config.Get()

	scrap, err := scraper.Start(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = scrap.Run(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
