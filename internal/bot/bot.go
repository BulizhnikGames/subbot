package bot

import (
	"context"
	"database/sql"
	"github.com/BulizhnikGames/subbot/db/orm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	commands map[string]CommandFunc
	db       *orm.Queries
	timeout  time.Duration
}

type CommandFunc func(ctx context.Context, bot *tgbotapi.BotAPI, db *orm.Queries, update tgbotapi.Update) error

func Start(token string, dburl string, timeout time.Duration) (*Bot, error) {
	dbConn, err := sql.Open("postgres", dburl)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	return &Bot{
		api:      bot,
		commands: make(map[string]CommandFunc),
		db:       orm.New(dbConn),
		timeout:  timeout,
	}, nil
}

func (b *Bot) WaitForUpdate(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case update := <-updates:
			updateCtx, cancel := context.WithTimeout(ctx, b.timeout)
			b.HandleUpdate(updateCtx, update)
			cancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bot) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	msgCommand := update.Message.Command()
	cmd, ok := b.commands[msgCommand]
	if !ok {
		return
	}

	if err := cmd(ctx, b.api, b.db, update); err != nil {
		log.Printf("Failed to exec command: %v", err)

		if _, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error")); err != nil {
			log.Printf("Failed to send error message: %v", err)
		}
	}
}
