package bot

import (
	"bufio"
	"context"
	"github.com/BulizhnikGames/subbot/internal/config"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"log"
	"os"
	"strings"
)

type Scraper struct {
	client *telegram.Client
	gaps   *updates.Manager
	Bot    *Bot
}

type Channel struct {
	ChannelID string `json:"ChannelID"`
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
		channelIDIndex := strings.Index(msg.String(), "ChannelID:")
		closeBracketIndex := strings.Index(msg.String(), "}")
		messageIDIndex := strings.Index(msg.String(), " ID:")
		fromIDIndex := strings.Index(msg.String(), " FromID:")
		if channelIDIndex < 0 || messageIDIndex < 0 || fromIDIndex < 0 || closeBracketIndex < 0 {
			return errors.New("unexpected message")
		}
		message := &Message{
			ChannelID: msg.String()[channelIDIndex+10 : closeBracketIndex],
			MessageID: msg.String()[messageIDIndex+4 : fromIDIndex],
		}
		//log.Printf("Got message from %s: %s", message.ChannelID, message.MessageID)
		messagesBuffer <- message
		return nil
	})

	return &Scraper{
		client: client,
		gaps:   gaps,
	}, nil
}

func (s *Scraper) Run(cfg config.Config) error {
	codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		log.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(code), nil
	}

	flow := auth.NewFlow(
		auth.Constant(
			cfg.Phone,
			cfg.Password,
			auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{})

	return s.client.Run(context.Background(), func(ctx context.Context) error {
		if err := s.client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		/*if testSub != "" {
			res, err := s.client.API().ContactsResolveUsername(ctx, testSub)
			if err == nil {
				channelID := res.Chats[0].GetID()
				var channel tg.InputChannel
				channel.ChannelID = channelID
				upd, err := s.client.API().ChannelsJoinChannel(ctx, &channel)
				if err != nil {
					log.Printf("Error subing to test: %v", err)
				} else {
					log.Printf("Subing to test completed: %s", upd.String())
				}
			}
		}*/

		user, err := s.client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}

		log.Printf("Scraper is %s:", user.Username)

		return s.gaps.Run(ctx, s.client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Println("Gaps started")
			},
		})
	})
}
