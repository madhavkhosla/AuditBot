package main

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	JiraUserName string
	JiraPassword string
	JiraBaseUrl string
	SlackClientId string
	SlackSecret string
	DatabaseUserName string
	DatabasePassword string
	DatabaseName string
	QuestionsFilePath string
	OAuthRedirectUri string
)

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Errorf("Fatal error config file: %s \n", err)
	} else {
		JiraUserName = viper.GetString("jira.username")
		JiraPassword = viper.GetString("jira.password")
		JiraBaseUrl = viper.GetString("jira.baseUrl")

		SlackClientId = viper.GetString("slack.clientId")
		SlackSecret = viper.GetString("slack.clientSecret")

		DatabaseName = viper.GetString("database.dbName")
		DatabaseUserName = viper.GetString("database.user")
		DatabasePassword = viper.GetString("database.password")

		QuestionsFilePath = viper.GetString("questions.path")

		OAuthRedirectUri = viper.GetString("oauth.redirectUri")
	}
}
