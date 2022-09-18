package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/zinstack625/mostful_manager/config"
	"github.com/zinstack625/mostful_manager/database"
	"github.com/zinstack625/mostful_manager/utils"
)

func (b *Bot) checkme(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
	}
	if req.Form.Get("token") != config.IntegrationTokens.CheckMe {
		resp.WriteHeader(403)
		resp.Write([]byte("Wrong token secret"))
	}
	labUrl := req.Form.Get("text")
	if ok, err := regexp.Match("https://github.com/.*/[0-9]{2}-lab-[0-9]{2}.*", []byte(labUrl)); err == nil && !ok {
		utils.RespondEphemeral(resp, "Does not seem like a lab we check! Make sure the URL is in form of \"https://github.com/bmstu-cbeer-20**/**-lab-**-YourName\"")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	student := &database.Student{
		MmstID:   req.Form.Get("user_id"),
		Tag:      req.Form.Get("user_name"),
		Labs:     []*database.Lab{},
		DoneLabs: []*database.DoneLab{},
	}
	database.DB.AddStudent(ctx, student)

	labNumScanner, err := regexp.Compile("https://github.com/.*/[0-9]{2}-lab-([0-9]{2}).*")
	if err != nil {
		log.Printf("Unable to compile regexp: %s", err)
		return
	}
	var labNum int64
	// Already checked that this will match, do not fret
	labNumString := string(labNumScanner.FindSubmatch([]byte(labUrl))[1])
	fmt.Sscanf(labNumString, "%d", &labNum)
	lab := database.Lab{
		Url:       labUrl,
		StudentID: student.ID,
		Number:    labNum,
	}
	existing, err := database.DB.GetLabs(ctx, &lab)
	if err != nil {
		log.Println("Unable to connect to database?: ", err.Error())
	}
	if len(existing) > 0 {
		utils.RespondEphemeral(resp, "Lab already added")
		return
	}
	mentor, _ := database.DB.AddLab(ctx, &lab)
	text := fmt.Sprintf("Lab %s, assigned to @%s", lab.Url, mentor.Tag)
	defer utils.RespondEphemeral(resp, text)
	go utils.SendDM(b.user.Id, req.Form.Get("user_id"), text, nil, b.client)
	mentor_msg := fmt.Sprintf("@%s: %s", student.Tag, lab.Url)

	action := model.PostAction{
		Id:   "approve",
		Type: "button",
		Name: "Approve",
		Integration: &model.PostActionIntegration{
			URL: "https://zinstack.ru/mmtest/actions",
			Context: map[string]interface{}{
				"action": map[string]interface{}{
					"type": "approve",
					"lab":  lab.ID,
				},
			},
		},
	}
	attachment := model.SlackAttachment{
		Actions: []*model.PostAction{&action},
	}

	go utils.SendDM(b.user.Id, mentor.MmstID, mentor_msg, []*model.SlackAttachment{&attachment}, b.client)
}

func (b *Bot) addmentor(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
	}
	if req.Form.Get("token") != config.IntegrationTokens.AddMentor {
		resp.WriteHeader(403)
		resp.Write([]byte("Wrong token secret"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if ok, err := database.DB.CheckAdmin(ctx, &database.Admin{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	}); err != nil || !ok {
		utils.RespondEphemeral(resp, "You have no permission!")
		return
	}
	args := strings.Split(req.Form.Get("text"), " ")
	if len(args) < 2 {
		utils.RespondEphemeral(resp, "Must supply Mattermost ID and Tag separated by space!")
		return
	}
	database.DB.AddMentor(ctx, &database.Mentor{
		MmstID: args[0],
		Tag:    args[1],
	})
	utils.RespondEphemeral(resp, "Done!")
}

func (b *Bot) removementor(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	if req.Form.Get("token") != config.IntegrationTokens.RemoveMentor {
		resp.WriteHeader(403)
		resp.Write([]byte("Wrong token secret"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if ok, err := database.DB.CheckAdmin(ctx, &database.Admin{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	}); err != nil || !ok {
		utils.RespondEphemeral(resp, "You have no permission!")
		return
	}
	args := strings.Split(req.Form.Get("text"), " ")
	if len(args) < 2 {
		utils.RespondEphemeral(resp, "Must supply Mattermost ID and Tag separated by space!")
		return
	}
	database.DB.RemoveMentor(ctx, &database.Mentor{
		MmstID: args[0],
		Tag:    args[1],
		Load:   0,
	})
	utils.RespondEphemeral(resp, "Done!")
}

func (b *Bot) myLabs(resp http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stud, err := database.DB.GetStudentByTag(ctx, req.Form.Get("user_id"))
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	table := make([][]string, 1)
	table[0] = make([]string, 13)
	if stud.RealName == nil {
		table[0][0] = stud.Tag
	} else {
		table[0][0] = *stud.RealName
	}
	for _, done_lab := range stud.Labs {
		table[0][done_lab.Number+1] = "✅"
	}
	utils.RespondEphemeral(resp, createMDTable(table))
}

func (b *Bot) labs(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	if req.Form.Get("token") != config.IntegrationTokens.Labs {
		resp.WriteHeader(403)
		resp.Write([]byte("Wrong token secret"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if ok, err := database.DB.CheckMentor(ctx, &database.Mentor{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	}); err != nil || !ok {
		b.myLabs(resp, req)
		return
	}
	studArray, err := database.DB.GetStudents(ctx)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	table := make([][]string, len(studArray))
	for i, v := range studArray {
		table[i] = make([]string, 13)
		if v.RealName == nil {
			table[i][0] = v.Tag
		} else {
			table[i][0] = *v.RealName
		}
		log.Printf("%p", v.DoneLabs)
		for _, done_lab := range studArray[i].Labs {
			table[i][done_lab.Number+1] = "✅"
		}
	}
	utils.RespondEphemeral(resp, createMDTable(table))
}

func createMDTable(table [][]string) string {
	var markdown string
	// HEADER
	markdown += " | "
	for i := range table[0] {
		markdown += fmt.Sprint(i + 1)
		if i == len(table[0])-2 {
			break
		}
		markdown += " | "
	}
	markdown += "\n"
	for i := range table[0] {
		markdown += "---"
		if i != len(table[0])-1 {
			markdown += " | "
		}
	}
	markdown += "\n"
	// BODY
	for _, row := range table {
		for i, column := range row {
			markdown += column
			if i != len(row)-1 {
				markdown += " | "
			}
		}
		markdown += "\n"
	}
	return markdown
}

func (b *Bot) SetupWebHooks() {
	http.HandleFunc("/checkme", b.checkme)
	http.HandleFunc("/addmentor", b.addmentor)
	http.HandleFunc("/removementor", b.removementor)
	http.HandleFunc("/actions", b.dispatchActions)
	http.HandleFunc("/labs", b.labs)
	go http.ListenAndServe("0.0.0.0:5000", nil)
}
