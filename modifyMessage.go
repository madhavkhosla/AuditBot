package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nlopes/slack"
)

func (a AuditBotClient) modifyForm(ev *slack.MessageEvent, userOpenFormMap map[string]*UserResource) bool {
	botId := a.Rtm.GetInfo().User.ID
	prefix := fmt.Sprintf("<@%s>", botId)
	if ev.User != botId && strings.HasPrefix(ev.Text, fmt.Sprintf("%s %s", prefix, MODIFY)) {
		// Modify command invoked by user to modify a particular question
		existingUserResource := userOpenFormMap[ev.User]
		if existingUserResource != nil {
			a.modifyMenu(ev, existingUserResource)
		} else {
			a.Rtm.SendMessage(a.Rtm.NewOutgoingMessage(
				fmt.Sprintf("User has not started the form yet"), ev.Channel))
		}
		return true
	}
	return false
}

func (a AuditBotClient) modifyMenu(ev *slack.MessageEvent, existingUserResource *UserResource) {

	questionOptions := []slack.AttachmentActionOption{}
	for i := 0; i < existingUserResource.lastAns; i++ {
		questionOptions = append(questionOptions, slack.AttachmentActionOption{Text: questions[i], Value: strconv.Itoa(i)})
	}

	attachment := slack.Attachment{
		Text:       "Choose the question to modify",
		Color:      "#7CD197",
		CallbackID: "modifyMenuCallbackId",
		Actions: []slack.AttachmentAction{
			{
				Name:    "Select",
				Type:    "select",
				Text:    "Modify Question",
				Options: questionOptions,
			},
		},
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
		AsUser: true,
	}

	if _, _, err := a.Rtm.PostMessage(ev.Channel, "", params); err != nil {
		a.err <- err
	}
}

func (a AuditBotClient) updateAnswer(ev *slack.MessageEvent, existingUserResource *UserResource) {
	inputStringLength := strings.Split(ev.Text, " ")
	modifyQuestionString := inputStringLength[3]
	modifyQuestion, err := strconv.Atoi(modifyQuestionString)
	modifyQuestion = modifyQuestion - 1
	if err != nil {
		a.err <- err
	}
	postMessgeParameters := slack.NewPostMessageParameters()
	postMessgeParameters.Attachments = []slack.Attachment{
		{
			Title: questions[modifyQuestion],
			Color: "#7CD197",
		},
	}
	a.Rtm.PostMessage(ev.Channel, "Please provide answer for question", postMessgeParameters)

	go a.modifyAnswerRoutine(modifyQuestion, existingUserResource, ev.Channel)
}

func (a AuditBotClient) modifyAnswerRoutine(modifyQuestion int, existingUserResource *UserResource, channel string) {

	modifyAnsEvent := <-existingUserResource.ModifyChannel
	fmt.Printf("Modify event text %s\n", modifyAnsEvent.Text)
	stmt, err := existingUserResource.DB.Prepare(fmt.Sprintf("UPDATE %s SET answer=? WHERE id=?", existingUserResource.FormName))
	if err != nil {
		a.err <- err
	}

	res, err := stmt.Exec(modifyAnsEvent.Text, modifyQuestion+1)
	if err != nil {
		panic(err)
		a.err <- err
	}
	id, err := res.RowsAffected()
	if err != nil {
		a.err <- err
	}
	fmt.Println(fmt.Sprintf("Last row inserted after modify %v", id))
	existingUserResource.Modify = false
	fmt.Printf("existingUserResource.lastAns %d\n", existingUserResource.lastAns)
	if existingUserResource.lastAns >= 3 {
		fmt.Println("Calling Submit again")
		a.submitForm(modifyAnsEvent, existingUserResource)
		return
	}
	if existingUserResource.lastAns >= 0 {
		existingUserResource.SyncChannel <- existingUserResource.lastAns
	}
}
