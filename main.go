package main

import (
	"net/http"

	"encoding/json"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"github.com/kr/pretty"
	"github.com/nlopes/slack"
	"github.com/pressly/chi"
	"io/ioutil"
	"net/url"
	"os"
)

var Handle apigatewayproxy.Handler

func init() {
	ln := net.Listen()
	Handle = apigatewayproxy.New(ln, nil).Handle

	r := chi.NewRouter()
	r.Post("/command", handleSlashCommand)
	r.Post("/action-endpoint", handleAction)
	go http.Serve(ln, r)
}

func handleSlashCommand(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.FormValue("team_domain") != os.Getenv("SLACK_TEAM") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.FormValue("token") != os.Getenv("VERIFICATION_TOKEN") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msg := slack.Msg{
		Text: "This is a test!",
		Attachments: []slack.Attachment{
			{
				CallbackID: "hogehoge",
				Actions: []slack.AttachmentAction{
					{
						Name: "ok",
						Text: "OK",
						Type: "button",
					},
					{
						Name: "cancel",
						Text: "Cancel",
						Type: "button",
					},
				},
			},
		},
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&msg)
}

func handleAction(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var payload slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pretty.Log(payload, r.Header)

	if payload.Team.Domain != os.Getenv("SLACK_TEAM") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if payload.Token != os.Getenv("VERIFICATION_TOKEN") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	msg := payload.OriginalMessage
	switch payload.Actions[0].Name {
	case "ok":
		msg.Text = "OK!!!"
	case "cancel":
		msg.Text = "Canceled..."
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&msg)
}
