package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/fujiwara/ridge"
)

const (
	urlFormat   = "https://access.line.me/dialog/oauth/weblogin?%s"
	redirectURL = "https://fyq3zlihs6.execute-api.ap-northeast-1.amazonaws.com/test/callback"
)

var (
	pageTempl = template.Must(template.ParseFiles("./templates/campaign.html"))
)

func callback(w http.ResponseWriter, r *http.Request) {
	log.Println("callback")
}

func showPage(w http.ResponseWriter, r *http.Request) {

	state := genState()

	u := url.Values{}
	u.Add("response_type", "code")
	u.Add("client_id", os.Getenv("WEB_CHANNEL_ID"))
	u.Add("redirect_uri", redirectURL)
	u.Add("state", state)

	url := fmt.Sprintf(urlFormat, u.Encode())

	surl := os.Getenv("WEB_STATIC_BASE_URL")

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	pageTempl.Execute(w, struct {
		LoginURL  string
		StaticURL string
	}{LoginURL: url, StaticURL: surl})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", callback)
	mux.HandleFunc("/campaign", showPage)

	ridge.Run(":8080", "", mux)
}
