package bot

import (
	"context"
	"github.com/BulizhnikGames/subbot/internal/config"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"log"
)

type Scraper struct {
	client *telegram.Client
	gaps   *updates.Manager
}

func StartScraper(apiID int, apiHash string) (*Scraper, error) {
	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})
	client := telegram.NewClient(apiID, apiHash, telegram.Options{UpdateHandler: gaps})
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.AsNotEmpty()
		if !ok {
			return errors.New("unexpected message")
		}
		log.Printf("Got message from %s channel: %s", msg.GetPeerID().String(), update.Message)
		return nil
	})

	return &Scraper{
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
