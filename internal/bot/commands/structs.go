package commands

import (
	"context"
	"github.com/BulizhnikGames/subbot/db/orm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type UserExpect struct {
	ExpectNext map[int64]CommandFunc
	Mutex      sync.Mutex
}

type CommandArguments struct {
	Api       *tgbotapi.BotAPI
	DB        *orm.Queries
	Status    *UserExpect
	UserID    int64
	GroupID   int64
	ChannelID int64
}

type CommandFunc func(ctx context.Context, args *CommandArguments) error
