package commands

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/db/orm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"sync"
)

type channelUsernameResult struct {
	username string
	err      error
}

func SlashList(ctx context.Context, args *CommandArguments) error {
	args.Status.Mutex.Lock()
	args.Status.ExpectNext[args.UserID] = nil
	args.Status.Mutex.Unlock()
	if err := writeListForGroup(ctx, args.Api, args.DB, args.GroupID); err != nil {
		return err
	}
	return nil
}

func writeListForGroup(ctx context.Context, bot *tgbotapi.BotAPI, db *orm.Queries, groupID int64) error {
	channels, err := db.ListGroupSubs(ctx, groupID)
	if err != nil {
		return err
	}
	var sb strings.Builder
	var wg sync.WaitGroup
	ch := make(chan *channelUsernameResult)
	for _, channel := range channels {
		go getChannelName(bot, channel, ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	sb.WriteString("Группа подписана на: ")
	for channel := range ch {
		if channel.err != nil {
			log.Printf("Can't get channel with ID %v: %v", groupID, channel.err)
		} else {
			sb.WriteString("@" + channel.username + " ")
		}
	}
	_, err = bot.Send(tgbotapi.NewMessage(groupID, sb.String()))
	return err
}

func getChannelName(bot *tgbotapi.BotAPI, channelID int64, ch chan<- *channelUsernameResult, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	cfg := tgbotapi.ChatInfoConfig{}
	cfg.ChatID = channelID
	chat, err := bot.GetChat(cfg)
	if err != nil {
		ch <- &channelUsernameResult{err: err}
		return
	}
	if !chat.IsChannel() {
		ch <- &channelUsernameResult{err: errors.New("not a channel")}
	}
	ch <- &channelUsernameResult{username: chat.UserName, err: nil}
}
