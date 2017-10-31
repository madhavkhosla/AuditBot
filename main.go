package main

import (
	"fmt"

	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	//"github.com/nlopes/slack"

	"log"
	"os"
	//"encoding/json"
	"encoding/json"

	"net/url"

	"strings"

	"github.com/nlopes/slack"
)

const CREATE = "create"

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	Bot         struct {
		BotUserId      string `json:"bot_user_id"`
		BotAccessToken string `json:"bot_access_token"`
	} `json:"bot"`
}

type SlackApp struct {
	ClientId     string
	ClientSecret string
}
type InteractiveMessageRequest struct {
	Actions []slack.AttachmentAction
	Channel slack.Channel
	User    slack.User
}

func (slackApp SlackApp) Submit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := os.Getenv("SLACK_AUTH_TOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Errorf("%s", err)
	}
	unEscapedBody, err := url.QueryUnescape(string(body))
	if err != nil {
		fmt.Errorf("%s", err)
	}
	i := strings.Index(unEscapedBody, "=")
	payload := unEscapedBody[i+1:]

	interactiveMessageRequest := InteractiveMessageRequest{}
	err = json.Unmarshal([]byte(payload), &interactiveMessageRequest)
	if err != nil {
		fmt.Printf("Error while un-marshaling request %s \n", err.Error())
	}
	postMessgeParameters := slack.NewPostMessageParameters()
	rtm.PostMessage(interactiveMessageRequest.Channel.ID,
		fmt.Sprintf("Submitted Form %s", interactiveMessageRequest.User.ID), postMessgeParameters)
	fmt.Println(interactiveMessageRequest.Actions[0].Value)
}

func (slackApp SlackApp) Auth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	code := r.URL.Query()["code"][0]

	url := fmt.Sprintf("https://slack.com/api/oauth.access?client_id=%s&client_secret=%s&code=%s&redirect_uri=http://localhost:8080/", "189197742244.254603813941", "8133da8b1cea1cc2d3647925c36d532e", code)
	authReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error during authetication %s", err.Error())
	}
	client := http.DefaultClient
	authResp, err := client.Do(authReq)
	if err != nil {
		fmt.Printf("Error during authetication %s", err.Error())
	}

	body, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		fmt.Printf("Error during authetication %s", err.Error())
	}

	authResponse := AuthResponse{}
	if err := json.Unmarshal(body, &authResponse); err != nil {
		fmt.Printf("Error during authetication %s", err.Error())
	}
	os.Setenv("SLACK_AUTH_TOKEN", authResponse.Bot.BotAccessToken)
	api := slack.New(authResponse.Bot.BotAccessToken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	go MessageLoop(rtm)

}

func main() {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	slackApp := SlackApp{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
	router := httprouter.New()
	router.GET("/", slackApp.Auth)
	router.POST("/submit", slackApp.Submit)

	log.Fatal(http.ListenAndServe(":8080", router))

}
