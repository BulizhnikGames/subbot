package scraper

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"github.com/BulizhnikGames/subbot/internal/bot"
	"github.com/BulizhnikGames/subbot/internal/config"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"log"
	"os"
	"strings"
	"time"
)

type Scraper struct {
	bot *bot.Bot
	client *telegram.Client
}

func Start(cfg config.Config) (*Scraper, error) {
	tgbot, err := bot.Start(cfg.Bot_token, cfg.DB_URL, 10 * time.Second)
	if err != nil{
		return nil, err
	}
	return &Scraper{
		bot: tgbot,
		client: telegram.NewClient(cfg.API_ID, cfg.API_hash, telegram.Options{})
	}, nil
}

func (s *Scraper) Run(cfg config.Config){
	err := s.client.Run(context.Background(), func(ctx context.Context) error {
		flow := auth.NewFlow(
			auth.Constant(cfg.Phone, cfg.Password, auth.Code())
		auth.SendCodeOptions{},
	)

		if err := s.client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}
		// Создаем обработчик обновлений
		handler := update.NewHandler()
		handler.OnNewChannelMessage(func(ctx context.Context, e *tg.UpdateNewChannelMessage) error {
			msg, ok := e.Message.(*tg.Message)
			if !ok {
				return nil
			}

			// Проверяем, что сообщение из нашего канала
			peerChannel, ok := msg.PeerID.(*tg.PeerChannel)
			if !ok {
				return nil
			}

			if peerChannel.ChannelID == channel.ID {
				fmt.Printf("Новое сообщение в канале %s: %s\n", channelUsername, msg.Message)
			}

			return nil
		})

		// Добавляем обработчик к клиенту
		s.client.AddEventHandler(handler)

		// Запускаем бесконечный цикл для прослушивания обновлений
		select {}
	})

	if err != nil {
		log.Fatal(err)
	}
}
