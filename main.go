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

	"strconv"

	"github.com/andygrunwald/go-jira"
	"github.com/nlopes/slack"
	"bufio"
)

const CREATE = "create"
const MODIFY = "modify"


type question struct {
	Text string
	Hint string
}

var questions []string

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
	if interactiveMessageRequest.Actions[0].Name == "Submit" {
		jiraIssueCreated := slackApp.createJiraIssue(interactiveMessageRequest)
		postMessgeParameters := slack.NewPostMessageParameters()
		var msg string
		if jiraIssueCreated {
			msg = fmt.Sprintf("Submitted Form %s", interactiveMessageRequest.User.ID)
		} else {
			msg = fmt.Sprintf("Error while subitting jira issue. Please press submit again." +
				"If error persists call the support team.")
		}
		rtm.PostMessage(interactiveMessageRequest.Channel.ID, msg, postMessgeParameters)
		fmt.Println(interactiveMessageRequest)
	} else if interactiveMessageRequest.Actions[0].Name == "Select" {

		questionToModify := interactiveMessageRequest.Actions[0].SelectedOptions[0].Value
		fmt.Printf("Question To Modify %s \n", questionToModify)
		questionNumber, err := strconv.Atoi(questionToModify)
		if err != nil {
			log.Println(err.Error())
		}
		userName := interactiveMessageRequest.User.ID
		postMessgeParameters := slack.NewPostMessageParameters()

		rtm.PostMessage(interactiveMessageRequest.Channel.ID,
			fmt.Sprintf("%s Modify Question %d", userName, questionNumber+1), postMessgeParameters)
	}
}

func (slackApp SlackApp) Auth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	code := r.URL.Query()["code"][0]

	url := fmt.Sprintf("https://slack.com/api/oauth.access?client_id=%s&client_secret=%s&code=%s&redirect_uri=%s", slackApp.ClientId, slackApp.ClientSecret, code, OAuthRedirectUri)
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

func (slackApp SlackApp) createJiraIssue(interactiveMessageRequest InteractiveMessageRequest) bool {
	var description string
	answers := strings.Split(interactiveMessageRequest.Actions[0].Value, ",")
	for i := 3; i < len(questions); i++ {
		description = fmt.Sprintf("%s\nQuestion; %s\nAnswer: %s", description, questions[i], answers[i])
	}
	jiraClient, err := jira.NewClient(nil, JiraBaseUrl)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	res, err := jiraClient.Authentication.AcquireSessionCookie(JiraUserName, JiraPassword)
	if err != nil || res == false {
		fmt.Printf("Result: %v\n", res)
		fmt.Println(err.Error())
	}
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				Name: "Automatic",
			},
			Description: description,
			Type: jira.IssueType{
				Name: "Story",
			},
			Project: jira.Project{
				Key: JiraProject,
			},
			Summary: fmt.Sprintf("Intake Form - %s", answers[0]),
			Reporter: &jira.User{
				Name: fmt.Sprintf("%s", answers[1]),
			},
			Duedate: answers[2],
		},
	}
	issue, _, err := jiraClient.Issue.Create(&i)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Println(issue)
	return true
}

func readQuestions(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Errorf("Error opening Questions file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	questionsSlice := make([]string,0)
	for scanner.Scan() {

		questionsSlice = append(questionsSlice, scanner.Text())
	}
	return questionsSlice
}

func main() {
	readConfig()
	questions = readQuestions(QuestionsFilePath)
	clientId := SlackClientId
	clientSecret := SlackSecret
	slackApp := SlackApp{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
	fmt.Println(slackApp)
	router := httprouter.New()
	router.GET("/", slackApp.Auth)
	router.POST("/submit", slackApp.Submit)

	log.Fatal(http.ListenAndServe(":8080", router))

}
