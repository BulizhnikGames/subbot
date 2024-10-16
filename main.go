package main

import (
	"github.com/BulizhnikGames/subbot/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	config.Load()
	cfg := config.Get()

}
