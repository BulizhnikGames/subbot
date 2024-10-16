package scraper

import (
	"context"
	"github.com/BulizhnikGames/subbot/internal/bot"
	"github.com/BulizhnikGames/subbot/internal/config"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"log"
	"time"
)

type Scraper struct {
	bot    *bot.Bot
	client *telegram.Client
	gaps   *updates.Manager
}

func Start(cfg config.Config) (*Scraper, error) {
	tgbot, err := bot.Start(cfg.Bot_token, cfg.DB_URL, 10*time.Second)
	if err != nil {
		return nil, err
	}

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})
	client := telegram.NewClient(cfg.API_ID, cfg.API_hash, telegram.Options{UpdateHandler: gaps})
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		log.Printf("Got message from %s channel: %s", e.Channels[0], update.Message)
		return nil
	})

	return &Scraper{
		bot:    tgbot,
		client: client,
		gaps:   gaps,
	}, nil
}

func (s *Scraper) Run(cfg config.Config) error {
	return s.client.Run(context.Background(), func(ctx context.Context) error {
		// Perform auth if no session is available.
		if _, err := s.client.Auth().Bot(ctx, cfg.Bot_token); err != nil {
			return err
		}

		user, err := s.client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}

		return s.gaps.Run(ctx, s.client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Println("Gaps started")
			},
		})
	})
}
