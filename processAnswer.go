package main

import (
	"fmt"

	"strings"

	"github.com/nlopes/slack"
)

func (a AuditBotClient) processAnswer(ev *slack.MessageEvent, userOpenFormMap map[string]*UserResource,
	userAllResourceMap map[string]map[string]*UserResource) {
	if ev.User != a.Rtm.GetInfo().User.ID && len(ev.User) > 0 {
		fmt.Println(ev.Text)
		existingUserResource, ok := userOpenFormMap[ev.User]
		if ok {
			if existingUserResource.Modify {
				existingUserResource.ModifyChannel <- ev
				return
			}
			existingUserResource.UserChannel <- ev
		}
	} else {
		if strings.Contains(ev.Text, "Submitted Form") {
			fmt.Println("SUMBIT CONDITION MET")
			inputStringLength := strings.Split(ev.Text, " ")
			user := inputStringLength[2]
			fmt.Println(user)
			fmt.Println(userOpenFormMap)
			fmt.Println(userAllResourceMap)
			formName := userOpenFormMap[user].FormName
			existingUserResource := userAllResourceMap[user]

			_, err := existingUserResource[formName].DB.Query(fmt.Sprintf("DROP TABLE %s", formName))
			if err != nil {
				a.err <- err
			}
			existingUserResource[formName].QuitChannel <- 0
			close(existingUserResource[formName].SyncChannel)
			close(existingUserResource[formName].UserChannel)
			close(existingUserResource[formName].QuitChannel)
			close(existingUserResource[formName].ModifyChannel)
			delete(existingUserResource, formName)
			delete(userOpenFormMap, user)
			fmt.Println(existingUserResource)
			fmt.Println(userOpenFormMap)
		} else if strings.Contains(ev.Text, "Modify Question") {
			// This is invoked by the formbot, to modify the question specified by user in the above step
			inputStringLength := strings.Split(ev.Text, " ")
			user := inputStringLength[0]
			fmt.Println(user)
			existingUserResource := userOpenFormMap[user]
			existingUserResource.Modify = true
			go a.updateAnswer(ev, existingUserResource)
		}
	}
}
