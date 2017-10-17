package main

import (
	"github.com/nlopes/slack"
	"fmt"
)

func (a AuditBotClient) HelpCommands(ev *slack.MessageEvent) bool {
	botId := a.Rtm.GetInfo().User.ID
	prefix := fmt.Sprintf("<@%s>", botId)
	if ev.User != botId && (ev.Text == prefix || ev.Text == fmt.Sprintf("%s help", prefix)) {
		postMessgeParameters := slack.NewPostMessageParameters()
		postMessgeParameters.Attachments = []slack.Attachment{
			{
				Title: "Command to start or restore an intake form",
				Text:  "@formbot create [Unique-id]",
				Color: "#7CD197",
			},
			{
				Title: "Command to modify a question once form is started",
				Text:  "@formbot modify",
				Color: "#7CD197",
			},
		}
		a.Rtm.PostMessage(ev.Channel, fmt.Sprintf("Formbot help commands"), postMessgeParameters)
		return true
	}
	return false
}