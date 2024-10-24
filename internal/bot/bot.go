package bot

import (
	"context"
	"database/sql"
	"github.com/BulizhnikGames/subbot/db/orm"
	"github.com/BulizhnikGames/subbot/internal/bot/commands"
	"github.com/BulizhnikGames/subbot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type Message struct {
	ChannelID int64
	MessageID int64
}

var messagesBuffer chan *Message

type Bot struct {
	api                *tgbotapi.BotAPI
	commands           map[string]commands.CommandFunc
	db                 *orm.Queries
	scraper            *Scraper
	expectNextFromUser *commands.UserExpect
	timeout            time.Duration
}

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

	return &Bot{
		api:      bot,
		commands: *RegisterCommands(),
		db:       orm.New(dbConn),
		scraper:  scraper,
		expectNextFromUser: &commands.UserExpect{
			ExpectNext: make(map[int64]commands.CommandFunc),
		},
		timeout: timeout,
	}, nil
}

func RegisterCommands() *map[string]commands.CommandFunc {
	cmds := make(map[string]commands.CommandFunc)
	cmds["list"] = commands.SlashList
	return &cmds
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

	if !update.Message.Chat.IsGroup() {
		if _, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команды этого бота можно использовать только в группах")); err != nil {
			log.Printf("Error sending restriction message (can use bot only in groups): %v", err)
		}
		return
	}

	msgCommand := update.Message.Command()
	cmd, ok := b.commands[msgCommand]
	if !ok {
		return
	}

	if err := cmd(ctx, &commands.CommandArguments{
		Api:       b.api,
		DB:        b.db,
		Status:    b.expectNextFromUser,
		UserID:    update.Message.From.ID,
		GroupID:   update.Message.Chat.ID,
		ChannelID: 0,
	}); err != nil {
		log.Printf("Failed to exec command: %v", err)

		if _, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error")); err != nil {
			log.Printf("Failed to send error message: %v", err)
		}
	}
}

func (b *Bot) SendNewPostsFromChannelToGroups(ctx context.Context) error {
	select {
	case msg := <-messagesBuffer:
		log.Printf("Sending post from %v to groups message with ID %v", msg.ChannelID, msg.MessageID)
		/*if _, err = b.api.Send(tgbotapi.NewForward(0, channelID, messageID)); err != nil {
			log.Printf("Failed to forward message from channel: %v",  err)
		}*/
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
