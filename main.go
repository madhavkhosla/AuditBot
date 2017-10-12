package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	"os"
)

func auth(clientId string) error {
	url := fmt.Sprintf("https://slack.com/oauth/authorize?client_id=%s&scope=bot&redirect_uri=http://localhost:8080/", "189197742244.254603813941")
	authReq, err := http.NewRequest("GET", url, nil)
	fmt.Println(authReq)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	authResp, err := client.Do(authReq)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func main() {
	clientId := os.Getenv("CLIENT_ID")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	go func() {
		http.ListenAndServe(":8080", nil)
	}()
	auth(clientId)
	//token := os.Getenv("SLACK_TOKEN")
	//api := slack.New(token)
	//rtm := api.NewRTM()
	//go rtm.ManageConnection()
	//go MessageLoop(rtm)
}
