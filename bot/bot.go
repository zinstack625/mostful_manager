package bot

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Bot struct {
	client   *model.Client4
	user     *model.User
	wsclient *model.WebSocketClient

	privatechannelid string
	debugchannelid   string
	ownUrl			 string
}

func (b *Bot) SetupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if b.wsclient != nil {
				b.wsclient.Close()
			}

			os.Exit(0)
		}
	}()
}

func (b *Bot) Init(instanceUrl, ownUrl, token, username, pchanID, dchanID string) {
	var err error
	b.wsclient, err = model.NewWebSocketClient4(fmt.Sprintf("wss://%s", instanceUrl), token)
	if err != nil {
		panic(err)
	}
	b.client = model.NewAPIv4Client(fmt.Sprintf("https://%s", instanceUrl))
	b.client.SetToken(token)
	b.user, _, err = b.client.GetUserByUsername(username, "")
	if err != nil {
		panic(err)
	}
	b.ownUrl = ownUrl
	b.privatechannelid = dchanID
	b.debugchannelid = pchanID
	b.wsclient.Listen()
	go func() {
		for resp := range b.wsclient.EventChannel {
			b.handleResp(resp)
		}
	}()
	go b.SetupWebHooks()
}

func (b *Bot) handleResp(resp *model.WebSocketEvent) {
	switch resp.EventType() {
	case "posted":
		b.dispatchPosted(resp)
	}
}

func (b *Bot) dispatchPosted(resp *model.WebSocketEvent) {
	switch resp.GetBroadcast().ChannelId {
	case b.privatechannelid:
		b.handleMentor(resp)
	case b.debugchannelid:
		return
	default:
		println(resp.GetBroadcast().ChannelId, resp.GetBroadcast().UserId)
		b.handleStudent(resp)
	}
}

func (b *Bot) handleMentor(resp *model.WebSocketEvent) {

}

func (b *Bot) handleStudent(resp *model.WebSocketEvent) {

}
