package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/zinstack625/mostful_manager/database"
)

func (b *Bot) checkme(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	student, err := database.DB.GetStudentByTag(ctx, req.Form.Get("user_id"))
	lab := database.Lab{
		Url: req.Form.Get("text"),
	}
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to connect to DB"))
		log.Println("Something went wrong with connecting to DB: ", err.Error())
	}
	if student == nil {
		student = &database.Student{
			Tag:      req.Form.Get("user_name"),
			Labs:     []*database.Lab{},
			DoneLabs: []*database.DoneLab{},
		}
	}
	student.Labs = append(student.Labs, &lab)
	if student == nil {
		database.DB.UpdateStudent(ctx, student)
	} else {
		database.DB.AddStudent(ctx, student)
	}
	lab.StudentID = student.ID
	mentor, _ := database.DB.AddLab(ctx, &lab)
	text := fmt.Sprintf("Ok, assigned to @%s", mentor.Tag)
	post := model.OutgoingWebhookResponse{
		Text:         &text,
		ResponseType: "ephemeral",
	}
	postjson, err := json.Marshal(post)
	if err != nil {
		panic(err)
	}
	resp.Write(postjson)
	dm, _, err := b.client.CreateDirectChannel(b.user.Id, req.Form.Get("user_id"))
	if err != nil {
		log.Println("Something went terribly wrong: ", err.Error())
	}
	postdmstud := model.Post{
		ChannelId: dm.Id,
		Message:   text,
		Props:     map[string]interface{}{},
	}
	b.client.CreatePost(&postdmstud)
}

func (b *Bot) SetupWebHooks() {
	testecho := func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "application/json")
		req.ParseForm()
		text := "echo echo echo"
		post := model.OutgoingWebhookResponse{
			Text:         &text,
			ResponseType: "ephemeral",
		}
		postjson, err := json.Marshal(post)
		if err != nil {
			panic(err)
		}
		resp.Write(postjson)
	}

	http.HandleFunc("/testecho", testecho)
	http.HandleFunc("/checkme", b.checkme)
	go http.ListenAndServe("0.0.0.0:80", nil)
}
