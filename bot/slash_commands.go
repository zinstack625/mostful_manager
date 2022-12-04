package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
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
	if ok, err := regexp.Match("^https://github.com/.*/[0-9]{2}-lab-[0-9]{2}.*/pull/[0-9]{1,}$", []byte(labUrl)); err == nil && !ok {
		utils.RespondEphemeral(resp, "Does not seem like a lab we check! Make sure the URL is in form of \"https://github.com/bmstu-cbeer-20**/**-lab-**-YourName/1\"")
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

	labNumScanner, err := regexp.Compile("^https://github.com/.*/[0-9]{2}-lab-([0-9]{2}).*/pull/[0-9]{1,}$")
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
	var report StudentsMarks
	report.students = make([]StudentReport, 1)
	studArray, err := database.DB.GetStudents(ctx)
	labNum := 0
	for _, stud := range studArray {
		studLabsCnt := len(stud.Labs) + len(stud.DoneLabs)
		if labNum < studLabsCnt {
			labNum = studLabsCnt
		}
	}
	report.students[0].labs = make([]LabState, labNum)
	for j := range report.students[0].labs {
		report.students[0].labs[j] = NotReady
	}
	if stud.RealName == nil {
		user, _, err := b.client.GetUsersByIds([]string{stud.MmstID})
		if err != nil && len(user) > 0 {
			report.students[0].name = user[0].GetFullName()
		} else {
			report.students[0].name = stud.Tag
		}
	} else {
		report.students[0].name = *stud.RealName
	}
	report.students[0].tag = fmt.Sprintf("@%s", stud.Tag)
	for _, done_lab := range stud.DoneLabs {
		report.students[0].labs[done_lab.Number] = Done
	}
	for _, sent_lab := range stud.Labs {
		report.students[0].labs[sent_lab.Number] = InProgress
	}
	utils.RespondEphemeral(resp, createMDTable(report))
}

type StudentsMarks struct {
	students        []StudentReport
	total_lab_count int
}

type LabState int

const (
	NotReady = iota
	InProgress
	Done
)

type StudentReport struct {
	name string
	tag  string
	labs []LabState
}

func (r *StudentsMarks) Len() int {
	return len(r.students)
}

func (r *StudentsMarks) Less(i, j int) bool {
	return r.students[i].name < r.students[j].name
}

func (r *StudentsMarks) Swap(i, j int) {
	r.students[i], r.students[j] = r.students[j], r.students[i]
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
	isMentor, err := database.DB.CheckMentor(ctx, &database.Mentor{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	})
	if err != nil {
		b.myLabs(resp, req)
		return
	}
	isAdmin, err := database.DB.CheckAdmin(ctx, &database.Admin{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	})
	if err != nil || !(isMentor || isAdmin) {
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
	var report StudentsMarks
	report.students = make([]StudentReport, len(studArray))
	report.total_lab_count = 0
	for _, stud := range studArray {
		studLabsCnt := len(stud.Labs) + len(stud.DoneLabs)
		if report.total_lab_count < studLabsCnt {
			report.total_lab_count = studLabsCnt
		}
	}
	for i, v := range studArray {
		report.students[i].labs = make([]LabState, report.total_lab_count)
		for j := range report.students[i].labs {
			report.students[i].labs[j] = NotReady
		}
		if v.RealName == nil {
			user, _, err := b.client.GetUsersByIds([]string{v.MmstID})
			if err != nil && len(user) > 0 {
				report.students[i].name = user[0].GetFullName()
			} else {
				report.students[i].name = v.Tag
			}
		} else {
			report.students[i].name = *v.RealName
		}
		report.students[i].tag = fmt.Sprintf("@%s", v.Tag)
		for _, done_lab := range studArray[i].DoneLabs {
			report.students[i].labs[done_lab.Number] = Done
		}
		for _, sent_lab := range studArray[i].Labs {
			report.students[i].labs[sent_lab.Number] = InProgress
		}
	}
	sort.Sort(&report)
	utils.RespondEphemeral(resp, createMDTable(report))
	if req.Form.Get("text") == "export" {
		channel, _, err := b.client.CreateDirectChannel(b.user.Id, req.Form.Get("user_id"))
		if err != nil {
			utils.RespondEphemeral(resp, "Unable to export!")
			return
		}
		file, _, err := b.client.UploadFile(makeCSV(report), channel.Id, "report.csv")
		if err != nil || len(file.FileInfos) == 0 {
			utils.RespondEphemeral(resp, "Unable to export!")
			return
		}
		post := model.PostEphemeral{
			UserID: channel.Id,
			Post: &model.Post{
				ChannelId: channel.Id,
				FileIds:   []string{file.FileInfos[0].Id},
			},
		}
		_, _, err = b.client.CreatePostEphemeral(&post)
		if err != nil {
			utils.RespondEphemeral(resp, "Unable to export!")
			return
		}
	}
}

func (b *Bot) mentorLabs(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	if req.Form.Get("token") != config.IntegrationTokens.MentorLabs {
		resp.WriteHeader(403)
		resp.Write([]byte("Wrong token secret"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	isAdmin, err := database.DB.CheckAdmin(ctx, &database.Admin{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	})
	if err != nil || !isAdmin {
		return
	}
	mentor, err := database.DB.GetMentorByTag(ctx, req.Form.Get("text"))
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to find mentor"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	stringBuffer := "Undone labs\n"
	for _, v := range mentor.Labs {
		stringBuffer += v.Url + "\n"
	}
	stringBuffer += "Done labs\n"
	for _, v := range mentor.DoneLabs {
		stringBuffer += v.Url + "\n"
	}
	utils.RespondEphemeral(resp, stringBuffer)
}

func createMDTable(table StudentsMarks) string {
	var markdown string
	// HEADER
	markdown += "Name | Tag | "
	for i := 0; i < table.total_lab_count-2; i++ {
		markdown += fmt.Sprint(i + 1)
		markdown += " | "
	}
	markdown += fmt.Sprintf("%d\n", table.total_lab_count)
	for i := 0; i < table.total_lab_count; i++ {
		markdown += "---"
		markdown += " | "
	}
	markdown += "---\n"
	// BODY
	for _, row := range table.students {
		markdown += fmt.Sprintf("%s | %s | ", row.name, row.tag)
		for i, column := range row.labs {
			switch column {
			case NotReady:
			case InProgress:
				markdown += "🔄"
			case Done:
				markdown += "✅"
			}
			if i != len(row.labs)-1 {
				markdown += " | "
			}
		}
		markdown += "\n"
	}
	return markdown
}

func makeCSV(table StudentsMarks) []byte {
	var csv string
	// HEADER
	csv += "Name,Tag,"
	for i := 0; i < table.total_lab_count-2; i++ {
		csv += fmt.Sprint(i + 1)
		csv += ","
	}
	csv += "\n"
	// BODY
	for _, row := range table.students {
		csv += fmt.Sprintf("%s | %s | ", row.name, row.tag)
		for i, column := range row.labs {
			switch column {
			case NotReady:
				csv += "0"
			case InProgress:
				csv += "1"
			case Done:
				csv += "2"
			}
			if i != len(row.labs)-1 {
				csv += ","
			}
		}
		csv += "\n"
	}
	return []byte(csv)
}

func (b *Bot) setStudName(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	err := req.ParseForm()
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("Unable to parse form"))
		log.Println("Something went wrong with parsing the url: ", err.Error())
		return
	}
	if req.Form.Get("token") != config.IntegrationTokens.SetName {
		resp.WriteHeader(403)
		resp.Write([]byte("Wrong token secret"))
	}
	args := strings.Split(req.Form.Get("text"), " ")
	if len(args) < 2 {
		utils.RespondEphemeral(resp, "Must supply Mattermost ID and Tag separated by space!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if ok, err := database.DB.CheckMentor(ctx, &database.Mentor{
		MmstID: req.Form.Get("user_id"),
		Tag:    req.Form.Get("user_name"),
	}); err != nil || !ok {
		utils.RespondEphemeral(resp, "You have no permission!")
		return
	}
	stud, err := database.DB.GetStudentByTag(ctx, args[0])
	if err != nil {
		log.Printf("Something went wrong at setting stud name, db.GetStudentByTag: %s", err)
		return
	}
	studName := args[1]
	for i := 2; i < len(args); i++ {
		studName += " " + args[i]
	}
	stud.RealName = &studName
	err = database.DB.UpdateStudent(ctx, stud)
	if err != nil {
		log.Printf("Something went wrong at setting stud name, db.UpdateStudent: %s", err)
		return
	}
	utils.RespondEphemeral(resp, "Done!")
}

func (b *Bot) SetupWebHooks() {
	http.HandleFunc("/checkme", b.checkme)
	http.HandleFunc("/addmentor", b.addmentor)
	http.HandleFunc("/removementor", b.removementor)
	http.HandleFunc("/actions", b.dispatchActions)
	http.HandleFunc("/labs", b.labs)
	http.HandleFunc("/setstudname", b.setStudName)
	http.HandleFunc("/mentorlabs", b.mentorLabs)
	go http.ListenAndServe("0.0.0.0:5000", nil)
}
