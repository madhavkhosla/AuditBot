package main

import (
	"github.com/nlopes/slack"
	"fmt"
	//"io/ioutil"
	"strconv"
)

var questions = []string{"Team Name", "PO Name", "Prod Release Date", "Business Justification"}

func (a AuditBotClient) sendQuestions(ev *slack.MessageEvent, syncChannel chan int,
	userAllResourceMap map[string]map[string]*UserResource, startI int, UniqueId string) {

	for i := startI; i < len(questions); {
		postMessgeParameters := slack.NewPostMessageParameters()
		postMessgeParameters.Attachments = []slack.Attachment{
			{
				Title: questions[i],
				Color: "#7CD197",
			},
		}
		a.Rtm.PostMessage(ev.Channel, "Question", postMessgeParameters)
		index := <-syncChannel
		if index == -1 {
			i = i + 1
		} else {
			fmt.Printf("Index is %v\n", index)
			i = index
		}
	}
	existingUserResource := userAllResourceMap[ev.User]
	fmt.Println(existingUserResource)
	a.submitForm(ev, existingUserResource[UniqueId])
}

func (a AuditBotClient) submitForm(ev *slack.MessageEvent, existingUserResource *UserResource) {


	rows, err := existingUserResource.DB.Query(fmt.Sprintf("SELECT answer FROM %s", existingUserResource.FormName))
	if err != nil {
		panic(err)
	}
	var allAnswers string = ""
	for rows.Next() {
		var answer string
		err = rows.Scan(&answer)
		if err != nil {
			panic(err)
		}
		if len(allAnswers) > 0 {
			allAnswers = fmt.Sprintf("%s,%s", allAnswers, answer)
		} else {
			allAnswers = fmt.Sprintf("%s", answer)
		}

	}



	//b, err := ioutil.ReadFile(fmt.Sprintf("/tmp/%s", existingUserResource.FormName))
	//if err != nil {
	//	panic(err)
	//}
	//ansFile := string(b)
	//fmt.Println(fmt.Sprintf("%s", ansFile))
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
