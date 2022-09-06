package utils

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
)

func GetMyIP() (net.Addr, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	return conn.LocalAddr(), err
}

func RespondEphemeral(resp http.ResponseWriter, text string) {
	post := model.OutgoingWebhookResponse{
		Text:         &text,
		ResponseType: "ephemeral",
	}
	postjson, _ := json.Marshal(post)
	resp.Write(postjson)
}

func SendDM(bot_id string, user_id string, msg string, attachments []*model.SlackAttachment, client *model.Client4) error {
	dm, _, err := client.CreateDirectChannel(bot_id, user_id)
	if err != nil {
		return err
	}
	postdmstud := model.Post{
		ChannelId: dm.Id,
		Message:   msg,
	}
	if attachments != nil {
		postdmstud.AddProp("attachments", attachments)
	}
	_, _, err = client.CreatePost(&postdmstud)
	return err
}
