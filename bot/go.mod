module github.com/BulizhnikGames/subbot/bot

go 1.23.0

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.7.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/BulizhnikGames/telegram-bot-api/v5 v5.5.3
