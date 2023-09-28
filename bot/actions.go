package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/zinstack625/mostful_manager/database"
)

type actionObject struct {
	Type              string `json:"type"`
	Lab               int    `json:"lab"`
	OriginalMessageID string
}

func (b *Bot) dispatchActions(resp http.ResponseWriter, req *http.Request) {
	log.Println("Got request...")
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	type contextMap struct {
		Action actionObject `json:"action"`
	}
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	var requestBody model.PostActionIntegrationRequest
	err = json.Unmarshal(body, &requestBody)
	actionCtx := actionObject{
		Type:              requestBody.Context["action"].(map[string]interface{})["type"].(string),
		Lab:               int(requestBody.Context["action"].(map[string]interface{})["lab"].(float64)),
		OriginalMessageID: requestBody.PostId,
	}

	dispatchMap := map[string]func(resp http.ResponseWriter, action *actionObject){
		"approve":    b.approveLab,
		"disapprove": b.disapproveLab,
	}

	if dispatchMap[actionCtx.Type] != nil {
		dispatchMap[actionCtx.Type](resp, &actionCtx)
	}
}

func (b *Bot) approveLab(resp http.ResponseWriter, action *actionObject) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	lab := database.Lab{
		ID: int64(action.Lab),
	}
	err := database.DB.GetLabPK(ctx, &lab)
	if err != nil {
		log.Printf("Something went wrong: %s", err)
		return
	}
	err = database.DB.FinishLab(ctx, &lab)
	if err != nil {
		log.Printf("Something went wrong at finishing: %s", err)
		return
	}

	postAction := model.PostAction{
		Id:   "disapprove",
		Type: "button",
		Name: "Disapprove",
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/actions", b.ownUrl),
			Context: map[string]interface{}{
				"action": map[string]interface{}{
					"type": "disapprove",
					"lab":  action.Lab,
				},
			},
		},
	}
	attachment := model.SlackAttachment{
		Actions: []*model.PostAction{&postAction},
	}
	op, _, _ := b.client.GetPost(context.Background(), action.OriginalMessageID, "")
	post := model.Post{
		Message: op.Message,
	}
	post.AddProp("attachments", []*model.SlackAttachment{&attachment})
	update := model.PostActionIntegrationResponse{
		Update:           &post,
		SkipSlackParsing: true,
	}
	updatejson, _ := json.Marshal(update)
	log.Println("Approving...")
	resp.Write(updatejson)
}

func (b *Bot) disapproveLab(resp http.ResponseWriter, action *actionObject) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	lab := database.DoneLab{
		ID: int64(action.Lab),
	}
	err := database.DB.GetDoneLabPK(ctx, &lab)
	if err != nil {
		log.Printf("Something went wrong: %s", err)
		return
	}
	err = database.DB.UnfinishLab(ctx, &lab)
	if err != nil {
		log.Printf("Something went wrong at Unfinishing: %s", err)
		return
	}
	postAction := model.PostAction{
		Id:   "approve",
		Type: "button",
		Name: "Approve",
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/actions", b.ownUrl),
			Context: map[string]interface{}{
				"action": map[string]interface{}{
					"type": "approve",
					"lab":  action.Lab,
				},
			},
		},
	}
	attachment := model.SlackAttachment{
		Actions: []*model.PostAction{&postAction},
	}
	op, _, _ := b.client.GetPost(context.Background(), action.OriginalMessageID, "")
	post := model.Post{
		Message: op.Message,
	}
	post.AddProp("attachments", []*model.SlackAttachment{&attachment})
	update := model.PostActionIntegrationResponse{
		Update:           &post,
		SkipSlackParsing: true,
	}
	updatejson, _ := json.Marshal(update)
	resp.Write(updatejson)
}

func (b *Bot) selfCheck(resp http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	err := database.DB.CheckConnection(ctx)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("db broken, fixme!"))
		log.Printf("THE SPY IS IN THE BASE! DATABASE BROKEN! FIXUP QUICK! %s", err.Error())
		return
	}
	resp.WriteHeader(200)
	resp.Write([]byte("imok"))
}
