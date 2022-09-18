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

func (b *Bot) Init(url string, token string) {
	var err error
	b.wsclient, err = model.NewWebSocketClient4(fmt.Sprintf("wss://%s", url), token)
	if err != nil {
		panic(err)
	}
	b.client = model.NewAPIv4Client(fmt.Sprintf("https://%s", url))
	b.client.SetToken(token)
	b.user, _, err = b.client.GetUserByUsername("cbeer_lab", "")
	if err != nil {
		panic(err)
	}
	b.privatechannelid = "xceb8h8ek3r6tbwmmynae7weba"
	b.debugchannelid = "7qttrxeyhbb1tei1qa9yn8remy"
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
