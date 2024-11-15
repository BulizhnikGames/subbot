package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (b *Bot) handleFromFetcher(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("message is not from fetcher: message empty")
	}

	if update.Message.ForwardFromChat == nil && update.Message.ForwardFrom == nil {
		return b.handleConfigMessage(update)
	}

	var chatID int64
	var err error
	if update.Message.ForwardFromChat != nil {
		chatID, err = tools.GetChannelID(update.Message.ForwardFromChat.ID)
		if err != nil {
			return err
		}
		log.Printf("Channel: ID: %v, Name: %s", chatID, update.Message.ForwardFromChat.UserName)
	} else {
		chatID = update.Message.ForwardFrom.ID
		log.Printf("User: ID: %v, Name: %s", chatID, update.Message.ForwardFrom.UserName)
	}

	messageID := update.Message.ForwardFromMessageID
	msgCfg := tools.MessageConfig{
		ChannelID: chatID,
		MessageID: messageID,
	}

	ok, err := b.tryHandleEdit(ctx, update, msgCfg, chatID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	ok, err = b.tryHandleRepost(ctx, update, msgCfg, chatID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	groups, err := b.db.GetSubsOfChannel(ctx, chatID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		b.tryUpdateChannelName(ctx, chatID, update.Message.ForwardFromChat.UserName)
		_, err := b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
		if err != nil {
			log.Printf("Error sending forward from channel %v to group %v: %v", chatID, group, err)
		}
	}
	return nil
}

func (b *Bot) handleConfigMessage(update tgbotapi.Update) error {
	if len(update.Message.Text) > 2 {
		if update.Message.Text[0] == 'r' { // got repost message config (ex: "r cID1 mID2 cID3 username")
			cfg, rep, err := tools.GetValuesFromRepostConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			b.channelReposts.Mutex.Lock()
			if b.channelReposts.List[*cfg] == nil {
				b.channelReposts.List[*cfg] = make([]tools.RepostedTo, 0)
			}
			b.channelReposts.List[*cfg] = append(b.channelReposts.List[*cfg], *rep)
			//log.Printf("added to reposts from: %+v to: %+v (%v)", *cfg, *rep, len(b.channelReposts.reposts[*cfg]))
			b.channelReposts.Mutex.Unlock()
			return nil
		} else if update.Message.Text[0] == 'e' { // got edit message config (ex: "e cID1 mID2 username")
			cfg, channelName, err := tools.GetValuesFromEditConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			b.channelEdit.Mutex.Lock()
			b.channelEdit.List[*cfg] = channelName
			b.channelEdit.Mutex.Unlock()
			//log.Printf("setted edit: %+v = %s", *cfg, channelName)
			return nil
		} else {
			return errors.New("message is not from fetcher: incorrect code in the beginning")
		}
	} else {
		return errors.New("message is not from fetcher: not forwarded and text length <= 2")
	}
}

func (b *Bot) tryHandleEdit(ctx context.Context, update tgbotapi.Update, msgCfg tools.MessageConfig, chatID int64) (bool, error) {
	b.channelEdit.Mutex.Lock()
	if channelName, ok := b.channelEdit.List[msgCfg]; ok {
		delete(b.channelEdit.List, msgCfg)
		b.channelEdit.Mutex.Unlock()

		groups, err := b.db.GetSubsOfChannel(ctx, chatID)
		if err != nil {
			return true, err
		}

		for _, group := range groups {
			_, err = b.api.Send(tgbotapi.NewMessage(group, "@"+channelName+" отредактировал сообщение:"))
			if err != nil {
				log.Printf("Error sending edited post from channel %v to group %v: %v", chatID, group, err)
				continue
			}

			_, err = b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
			if err != nil {
				log.Printf("Error sending edited post from channel %v to group %v: %v", chatID, group, err)
			}
		}
		return true, nil
	} else {
		b.channelEdit.Mutex.Unlock()
		return false, nil
	}
}

func (b *Bot) tryHandleRepost(ctx context.Context, update tgbotapi.Update, msgCfg tools.MessageConfig, chatID int64) (bool, error) {
	b.channelReposts.Mutex.Lock()
	if targets, ok := b.channelReposts.List[msgCfg]; ok {
		delete(b.channelReposts.List, msgCfg)
		b.channelReposts.Mutex.Unlock()

		for _, target := range targets {
			groups, err := b.db.GetSubsOfChannel(ctx, target.ChannelID)
			if err != nil {
				if len(targets) > 0 {
					log.Printf(
						"Could not repost message from channel %v, to channel (%v, %s): %v",
						chatID,
						target.ChannelID,
						target.ChannelName,
						err,
					)
					continue
				} else {
					return true, errors.New(
						fmt.Sprintf(
							"Could not repost message from channel %v, to channel (%v, %s): %v",
							chatID,
							target.ChannelID,
							target.ChannelName,
							err,
						),
					)
				}
			}

			for _, group := range groups {
				_, err = b.api.Send(tgbotapi.NewMessage(group, "@"+target.ChannelName+" переслал сообщение:"))
				if err != nil {
					log.Printf("Error sending repost from channel %v to group %v: %v", chatID, group, err)
					continue
				}

				_, err = b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
				if err != nil {
					log.Printf("Error sending repost from channel %v to group %v: %v", chatID, group, err)
				}
			}
		}
		return true, nil
	} else {
		b.channelReposts.Mutex.Unlock()
		return false, nil
	}
}