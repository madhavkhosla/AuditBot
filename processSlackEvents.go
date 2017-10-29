package main

import (
	"github.com/nlopes/slack"
	"fmt"

	"database/sql"
)

type AuditBotClient struct {
	Rtm *slack.RTM

}

type UserResource struct {
	UserChannel   chan *slack.MessageEvent
	ModifyChannel chan *slack.MessageEvent
	SyncChannel   chan int
	DB            *sql.DB
	QuitChannel   chan int
	FormName      string
	Modify        bool
	lastAns		  int
}

func MessageLoop(rtm *slack.RTM) {
	auditBotClient := AuditBotClient{rtm}
	userOpenFormMap := make(map[string]*UserResource)
	userAllResourceMap := make(map[string]map[string]*UserResource)

Loop:
	for {
		msg := <-rtm.IncomingEvents
		fmt.Println("Event Received: ")
		switch ev := msg.Data.(type) {

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change :%s %s \n", ev.User, ev.Presence)

		case *slack.ConnectedEvent:
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			fmt.Println(ev.Msg.BotID)
			fmt.Printf("Message: %v\n", ev.Msg.Text)

			// AuditBot help commands
			helpCommandInvoked := auditBotClient.HelpCommands(ev)
			if helpCommandInvoked {
				continue Loop
			}
			createCommandInvoked := auditBotClient.createMessage(ev, userOpenFormMap, userAllResourceMap)
			if createCommandInvoked {
				continue Loop
			}
			auditBotClient.processAnswer(ev, userOpenFormMap)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			break Loop

		default:
			fmt.Println(msg.Type)
		}

	}
}


