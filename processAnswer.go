package main

import (
	"fmt"
	"github.com/nlopes/slack"
)

func (a AuditBotClient) processAnswer(ev *slack.MessageEvent, userOpenFormMap map[string]*UserResource) {
	if ev.User != a.Rtm.GetInfo().User.ID && len(ev.User) > 0 {
		fmt.Println(ev.Text)
	existingUserResource, ok := userOpenFormMap[ev.User]
		if ok {
			existingUserResource.UserChannel <- ev
		}
	}

}