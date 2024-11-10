package fetcher

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/BulizhnikGames/subbot/fetcher/tools"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Fetcher struct {
	client  *telegram.Client
	gaps    *updates.Manager
	botID   int64
	botHash int64
}

func Init(apiID int, apiHash string, botUsername string) (*Fetcher, error) {
	f := &Fetcher{
		botID:   0,
		botHash: 0,
	}

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})
	session := telegram.FileSessionStorage{
		Path: "./session.json",
	}
	client := telegram.NewClient(apiID, apiHash, telegram.Options{UpdateHandler: gaps, SessionStorage: &session})
	// TODO: handle edit message
	/*d.OnEditChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateEditChannelMessage) error{

	})*/
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			log.Printf("Unexpected message type: %T", update.Message)
			return errors.New("unexpected message")
		}

		// If admin of channel A forwards post from channel B to his channel
		// (Or forwards old channel from his own channel (A), but for simplicity consider first variant)
		// All groups which are subs of channel A will get repost of B's post which will look strange
		// Because it will show B as the channel where it posted, but group might not be subscribed to B
		// (Technically it's correct, but group members might be confused why they get messages from random channels)
		// Also all groups which are subs of channel B will get it's old post which will be also strange
		// Because bot sends only new posts from each channel
		// TODO: handle this situation by sending to subbot message containing ID and username of channel A
		// TODO: so subbot will forward this message to valid groups and send explaining message before forward:
		// TODO: "channel @A forwarded message:"
		if _, ok := msg.GetFwdFrom(); ok {
			log.Printf("Message won't be forwarded, because it's already forwarded")
			return nil
		}

		peer, ok := msg.PeerID.(*tg.PeerChannel)
		if !ok {
			log.Printf("Unexpected message's peer type: %T", msg.PeerID)
			return errors.New("unexpected peer")
		}

		getChannel, err := client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: 0,
			},
		})
		if err != nil {
			log.Printf("Error getting channels access hash: %v", err)
			return err
		}
		channelData, ok := getChannel.(*tg.MessagesChats)
		if !ok {
			log.Printf("Unexpected channel type: %T", getChannel)
			return errors.New("unexpected channel")
		} else if channelData.Chats == nil {
			log.Printf("Error: empty channel")
			return errors.New("unexpected channel")
		}
		channel, ok := channelData.Chats[0].(*tg.Channel)
		if !ok {
			log.Printf("Unexpected channel chat type: %T", channelData.Chats[0])
			return errors.New("unexpected channel")
		}

		/*getBot, err := client.API().UsersGetFullUser(ctx, &tg.InputUser{
			UserID:     botChat,
			AccessHash: 0,
		})
		if err != nil {
			log.Println(err)
			return err
		}
		bot, ok := getBot.Chats[0].(*tg.Chat)
		if !ok {
			log.Println(err)
			return errors.New("unexpected bot")
		}

		log.Printf("%v, %v \n %v, %v", peer.ChannelID, channel.AccessHash, bot.ID)*/

		var botPeer tg.InputPeerClass
		if f.botID == 0 || f.botHash == 0 {
			resolved, err := client.API().ContactsResolveUsername(ctx, botUsername)
			if err != nil {
				log.Printf("failed to resolve username: %v", err)
				return err
			}

			if len(resolved.Users) > 0 {
				user := resolved.Users[0]
				if u, ok := user.(*tg.User); ok {
					botPeer = &tg.InputPeerUser{
						UserID:     u.ID,
						AccessHash: u.AccessHash,
					}
					f.botID = u.ID
					f.botHash = u.AccessHash
				}
			} else {
				return fmt.Errorf("could not resolve bot username")
			}
		} else {
			botPeer = &tg.InputPeerUser{
				UserID:     f.botID,
				AccessHash: f.botHash,
			}
		}

		//log.Printf("%v, %v", botPeer.(*tg.InputPeerUser).UserID, botPeer.(*tg.InputPeerUser).AccessHash)

		_, err = client.API().MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
			FromPeer: &tg.InputPeerChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: channel.AccessHash,
			},
			ToPeer:   botPeer,
			ID:       []int{msg.ID},
			RandomID: []int64{rand.Int63()},
		})
		if err != nil {
			log.Printf("Error forwarding messages: %v", err)
		}
		return err
	})

	f.client = client
	f.gaps = gaps
	return f, nil
}

func (f *Fetcher) Run(phone string, password string, apiURL string, IP string, port string) error {
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
			phone,
			password,
			auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{})

	return f.client.Run(context.Background(), func(ctx context.Context) error {
		if err := f.client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		user, err := f.client.Self(ctx)
		if err != nil {
			return err
		}

		jsonBody := []byte(fmt.Sprintf("{\"phone\": \"%s\", \"ip\": \"%s\", \"port\": \"%s\"}", phone, IP, port))
		bodyReader := bytes.NewReader(jsonBody)
		requestURL := apiURL + "/" + strconv.FormatInt(user.ID, 10)
		req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
		if err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("[POSSIBLE ERROR]: couldn't register fetcher by API request (register manualy or restart): %v", err)
		} else if res.StatusCode != http.StatusOK {
			log.Printf("[POSSIBLE ERROR]: couldn't register fetcher by API request (register manualy or restart). HTTP status code: %v", res.StatusCode)
		}

		log.Printf("Scraper is %s:", user.Username)

		return f.gaps.Run(ctx, f.client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Println("Gaps started")
			},
		})
	})
}

func (f *Fetcher) SubscribeToChannel(ctx context.Context, channelName string) (int64, int64, error) {
	channelID, accessHash, err := f.GetChannelInfo(ctx, channelName)
	if err != nil {
		return 0, 0, err
	}
	channel := tg.InputChannel{ChannelID: channelID, AccessHash: accessHash}
	_, err = f.client.API().ChannelsJoinChannel(ctx, &channel)
	return channelID, accessHash, err
}

func (f *Fetcher) GetChannelInfo(ctx context.Context, channelName string) (int64, int64, error) {
	res, err := f.client.API().ContactsResolveUsername(ctx, channelName)
	if err != nil {
		return 0, 0, err
	}
	channelID, err := tools.GetChannelIDFromChannel(res.Chats[0].String())
	if err != nil {
		return 0, 0, err
	}
	accessHash, err := tools.GetAccessHashFromChannel(res.Chats[0].String())
	if err != nil {
		return 0, 0, err
	}
	return channelID, accessHash, err
}
