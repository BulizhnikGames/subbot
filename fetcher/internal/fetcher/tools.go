package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
	"log"
	"math/rand"
	"time"
)

func (f *Fetcher) getChannelAndMessageInfo(ctx context.Context, message tg.MessageClass) (*tg.Channel, *tg.Message, error) {
	msg, ok := message.(*tg.Message)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("unexpected message type %T:", message))
	}

	peer, ok := msg.PeerID.(*tg.PeerChannel)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("unexpected peer type: %T", msg.PeerID))
	}

	getChannel, err := f.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{
			ChannelID:  peer.ChannelID,
			AccessHash: 0,
		},
	})
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Error getting channels (%v) access hash: %v", peer.ChannelID, err))
	}
	channelData, ok := getChannel.(*tg.MessagesChats)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("unexpected channel type %T", getChannel))
	} else if channelData.Chats == nil {
		return nil, nil, errors.New("unexpected channel: channel empty")
	}
	channel, ok := channelData.Chats[0].(*tg.Channel)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("unexpected channel chat type %T", channelData.Chats[0]))
	}
	return channel, msg, nil
}

func (f *Fetcher) setBotHashAndID(ctx context.Context) error {
	resolved, err := f.client.API().ContactsResolveUsername(ctx, f.botUsername)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to resolve username of bot (%v) to foward message: %v", f.botUsername, err))
	}

	if len(resolved.Users) > 0 {
		user := resolved.Users[0]
		if u, ok := user.(*tg.User); ok {
			f.botID = u.ID
			f.botHash = u.AccessHash
		} else {
			return errors.New(fmt.Sprintf("failed to resolve username of bot (%v): not a user", f.botUsername))
		}
	} else {
		return errors.New(fmt.Sprintf("failed to resolve username of bot (%v): resolving returned 0 users", f.botUsername))
	}

	return nil
}

func (f *Fetcher) waitForOtherMediaInGroup(groupID int64) {
	time.Sleep(f.mediaWaitTimer)
	f.multiMediaQueue.Mutex.Lock()
	defer f.multiMediaQueue.Mutex.Unlock()
	if f.multiMediaQueue.List[groupID] != nil {
		f.sendChan <- f.multiMediaQueue.List[groupID]
		delete(f.multiMediaQueue.List, groupID)
		log.Printf("Forwarded media group: %v", groupID)
	}
}

func getRandomIDs(n int) []int64 {
	res := make([]int64, n)
	for i := 0; i < n; i++ {
		res[i] = rand.Int63()
	}
	return res
}
