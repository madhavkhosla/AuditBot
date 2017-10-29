package main

import (
	"fmt"

	"strconv"

	"github.com/nlopes/slack"
	"database/sql"
)

var questions = []string{"Team Name", "PO Name", "Prod Release Date", "Business Justification"}

func (a AuditBotClient) sendQuestions(ev *slack.MessageEvent, syncChannel chan int,
	userAllResourceMap map[string]map[string]*UserResource, UniqueId string) {

	existingUserResource := userAllResourceMap[ev.User]
	currentOpenForm := existingUserResource[UniqueId]
	for {
		postMessgeParameters := slack.NewPostMessageParameters()
		postMessgeParameters.Attachments = []slack.Attachment{
			{
				Title: questions[currentOpenForm.lastAns],
				Color: "#7CD197",
			},
		}
		a.Rtm.PostMessage(ev.Channel, "Question", postMessgeParameters)
		index := <-syncChannel
		if index >= len(questions) {
			fmt.Println("last question answered")
			break
		} else {
			fmt.Printf("last row inserted %v\n", index)
			currentOpenForm.lastAns = index
		}
	}
	fmt.Println(existingUserResource)
	a.submitForm(ev, currentOpenForm)
}

func (a AuditBotClient) submitForm(ev *slack.MessageEvent, existingUserResource *UserResource) {

	allAnswers, _ := a.readTable(ev.Channel, existingUserResource.DB, existingUserResource.FormName)

	postMessgeParameters := slack.NewPostMessageParameters()
	postMessgeParameters.AsUser = true
	questionOptions := []slack.AttachmentActionOption{}

	for i, q := range questions {
		questionOptions = append(questionOptions, slack.AttachmentActionOption{Text: q, Value: strconv.Itoa(i)})
	}
	postMessgeParameters.Attachments = []slack.Attachment{
		{
			Title: "Do you want to submit the intake form",
			Color: "#7CD197",
			Actions: []slack.AttachmentAction{
				{
					Name:  "Submit",
					Text:  "Submit",
					Type:  "button",
					Value: fmt.Sprintf("%s", allAnswers),
				},
				{
					Name:    "Select",
					Type:    "select",
					Text:    "Modify Question",
					Options: questionOptions,
				},
			},
			CallbackID: "callbackId",
		},
	}
	a.Rtm.PostMessage(ev.Channel, "", postMessgeParameters)

}

func (a AuditBotClient) readTable(eventChannel string, db *sql.DB, formName string) (string, int) {
	answerArray := make([]slack.AttachmentField, 0, len(questions))
	rows, err := db.Query(fmt.Sprintf("SELECT answer FROM %s", formName))
	if err != nil {
		panic(err)
	}
	questionAnsweredCount := 0
	var allAnswers string = ""
	for rows.Next() {
		var answer string
		err = rows.Scan(&answer)
		if err != nil {
			panic(err)
		}
		answerArray = append(answerArray, slack.AttachmentField{
				Title: questions[questionAnsweredCount],
				Value: answer,
				Short: false,
			})
		if len(allAnswers) > 0 {
			allAnswers = fmt.Sprintf("%s,%s", allAnswers, answer)
		} else {
			allAnswers = fmt.Sprintf("%s", answer)
		}
		questionAnsweredCount += 1
	}
		postMessgeParameters := slack.NewPostMessageParameters()
		postMessgeParameters.AsUser = true
		postMessgeParameters.Attachments = []slack.Attachment{
			{
				Title:  "Intake form filled till now.",
				Color:  "#7CD197",
				Fields: answerArray,
			},
		}
		a.Rtm.PostMessage(eventChannel, "", postMessgeParameters)
	return allAnswers, questionAnsweredCount
}
