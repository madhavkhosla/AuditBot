package main

import (
	"github.com/nlopes/slack"
	"strings"
	"fmt"


	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)

func (a AuditBotClient) createMessage(ev *slack.MessageEvent, userOpenFormMap map[string]*UserResource, userAllResourceMap map[string]map[string]*UserResource) bool {
	botId := a.Rtm.GetInfo().User.ID
	prefix := fmt.Sprintf("<@%s>", botId)
	if ev.User != a.Rtm.GetInfo().User.ID && strings.HasPrefix(ev.Text, fmt.Sprintf("%s %s", prefix, CREATE)) {
		// Input create command not correct
		check := a.invalidCreateCommand(ev)
		if !check {
			return check
		}
		// Create command starts a form
		go a.startForm(ev, userOpenFormMap, userAllResourceMap)
		return true
	}
	return false
}

func (a AuditBotClient) invalidCreateCommand(ev *slack.MessageEvent) bool {
	inputStringLength := strings.Split(ev.Text, " ")
	if len(inputStringLength) != 3 {

		postMessgeParameters := slack.NewPostMessageParameters()
		postMessgeParameters.Attachments = []slack.Attachment{
			{
				Title: "Incorrect input command for creating intake form",
				Text:  "Please try the help command: @auditbot help",
				Color: "#7CD197",
			},
		}
		a.Rtm.PostMessage(ev.Channel, "", postMessgeParameters)

		return false
	}
	return true
}

func (a AuditBotClient) startForm(ev *slack.MessageEvent, userOpenFormMap map[string]*UserResource, userAllResourceMap map[string]map[string]*UserResource) {

	fmt.Println("Inside Start form")
	syncChannel := make(chan int)
	userChannel := make(chan *slack.MessageEvent)
	modifyChannel := make(chan *slack.MessageEvent)
	inputStringLength := strings.Split(ev.Text, " ")
	UniqueId := inputStringLength[2]
	_, ok := userAllResourceMap[ev.User][UniqueId]
	db, err := sql.Open("mysql", "madhav:password@/Auditbot")
	if err != nil {
		fmt.Errorf(err.Error())
	}

	if !ok {
		newUserFormResourceMap := make(map[string]*UserResource)

		formTableExistsStatement := fmt.Sprintf("show tables like '%s';", UniqueId)
		rows, err := db.Query(formTableExistsStatement)
		if err != nil {
			panic("DB connection failed")
		}
		if !rows.Next() {
			fmt.Println("Form table does not exist")
			// create form table
			_, err := db.Exec(fmt.Sprintf("CREATE TABLE %s ( id int(10) NOT NULL AUTO_INCREMENT, answer varchar(500),  PRIMARY KEY (id) )", UniqueId))
			if err != nil {
				panic(err)
			}

			newUserFormResourceMap[UniqueId] = &UserResource{userChannel,
				modifyChannel,
				syncChannel,
				db,
				make(chan int),
				UniqueId,
				false}
			userAllResourceMap[ev.User] = newUserFormResourceMap
			userOpenFormMap[ev.User] = userAllResourceMap[ev.User][UniqueId]
			fmt.Println(userAllResourceMap)
			go a.sendQuestions(ev, syncChannel, userAllResourceMap, 0, UniqueId)
		}
		go a.startUserRoutine(userOpenFormMap[ev.User])
	}
}

func (a AuditBotClient) startUserRoutine(existingUserResource *UserResource) {
	for {
		select {
		case userEvent := <-existingUserResource.UserChannel:

			stmt, err := existingUserResource.DB.Prepare(fmt.Sprintf("INSERT %s SET answer=?", existingUserResource.FormName))
			if err != nil {
				panic(err)
			}

			res, err := stmt.Exec(userEvent.Text)
			if err != nil {
				panic(err)
			}
			id, err := res.LastInsertId()
			if err != nil {
				panic(err)
			}
			fmt.Println(fmt.Sprintf("Last row inserted %v", id))
			existingUserResource.SyncChannel <- -1
		case <-existingUserResource.QuitChannel:
			fmt.Println("quit")
			return
		}
	}
}