package bot

import (
	"context"
	"database/sql"
	"github.com/BulizhnikGames/subbot/db/orm"
	"github.com/BulizhnikGames/subbot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type Message struct {
	ChannelID string
	MessageID string
}

type Bot struct {
	api      *tgbotapi.BotAPI
	commands map[string]CommandFunc
	db       *orm.Queries
	Scraper  *Scraper
	timeout  time.Duration
}

var messagesBuffer chan *Message

type CommandFunc func(ctx context.Context, bot *tgbotapi.BotAPI, db *orm.Queries, update tgbotapi.Update) error

func StartBot(cfg config.Config, timeout time.Duration) (*Bot, error) {
	messagesBuffer = make(chan *Message, 40)

	dbConn, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Bot_token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	scraper, err := StartScraper(cfg.API_ID, cfg.API_hash)
	if err != nil {
		return nil, err
	}

	res := &Bot{
		api:      bot,
		commands: make(map[string]CommandFunc),
		db:       orm.New(dbConn),
		Scraper:  scraper,
		timeout:  timeout,
	}

	scraper.Bot = res

	return res, nil
}

func (b *Bot) WaitForUpdate(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	log.Println("Waiting for commands...")

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

func (b *Bot) WaitForNewPosts(ctx context.Context) error {
	select {
	case msg := <-messagesBuffer:
		log.Printf("Sending post from %s to groups message with ID %s", msg.ChannelID, msg.MessageID)
		/*channelID, err := strconv.ParseInt(msg.ChannelID, 10, 64)
		if err != nil {
			log.Printf("Failed to convert channel id (%s) to int: %v", msg.ChannelID, err)
			//TODO: Send error message
		}
		messageID, err := strconv.Atoi(msg.MessageID)
		if err != nil {
			log.Printf("Failed to convert message id (%s) to int: %v", msg.MessageID, err)
			//TODO: Send error message
		}
		if _, err = b.api.Send(tgbotapi.NewForward(0, channelID, messageID)); err != nil {
			log.Printf("Failed to forward message from channel: %v",  err)
			//TODO: Send error message
		}*/
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
