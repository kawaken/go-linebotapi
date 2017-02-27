package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fujiwara/ridge"
)

const (
	urlFormat  = "https://access.line.me/dialog/oauth/weblogin?%s"
	grantURL   = "https://api.line.me/v2/oauth/accessToken"
	profileURL = "https://api.line.me/v2/profile"
)

var (
	pageTempl   = template.Must(template.ParseFiles("./templates/campaign.html"))
	couponTempl = template.Must(template.ParseFiles("./templates/coupon.html"))
)

type authzResponse struct {
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type profile struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

func grant(params url.Values) (*authzResponse, error) {
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", os.Getenv("WEB_CHANNEL_ID"))
	params.Add("client_secret", os.Getenv("WEB_CHANNEL_SECRET"))
	params.Add("redirect_uri", os.Getenv("WEB_REDIRECT_URL"))

	resp, err := http.PostForm(grantURL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ar authzResponse
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

func getProfile(token string) (*profile, error) {
	req, err := http.NewRequest(http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p profile
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func saveUser(ar *authzResponse, p *profile) error {
	svc, err := newService()
	if err != nil {
		return err
	}

	params := &dynamodb.PutItemInput{
		TableName: aws.String("users"),
		Item: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(p.UserID),
			},
			"display_name": {
				S: aws.String(p.DisplayName),
			},
			"access_token": {
				S: aws.String(ar.AccessToken),
			},
			"refresh_token": {
				S: aws.String(ar.RefreshToken),
			},
			"expired_in": {
				N: aws.String(strconv.Itoa(ar.ExpiresIn)),
			},
			"registered_at": {
				N: aws.String(strconv.FormatInt(time.Now().Unix(), 10)),
			},
			"hash": {
				S: aws.String(genState()),
			},
		},
	}

	// PutItemの実行
	_, err = svc.PutItem(params)
	if err != nil {
		return err
	}

	return nil
}

func setCoupon(userID string) error {
	svc, err := newService()
	if err != nil {
		return err
	}

	params := &dynamodb.PutItemInput{
		TableName: aws.String("coupons"),
		Item: map[string]*dynamodb.AttributeValue{
			"coupon_id": {
				S: aws.String("1"),
			},
			"user_id": {
				S: aws.String(userID),
			},
			"registered_at": {
				N: aws.String(strconv.FormatInt(time.Now().Unix(), 10)),
			},
		},
	}

	// PutItemの実行
	_, err = svc.PutItem(params)
	if err != nil {
		return err
	}

	return nil
}

func callback(w http.ResponseWriter, r *http.Request) {
	surl := os.Getenv("WEB_STATIC_BASE_URL")
	errURL := surl + "/error.html"
	sucURL := surl + "/success.html"

	params := r.URL.Query()
	if params.Get("error") != "" {
		http.Redirect(w, r, errURL, 301)
		return
	}

	params.Del("state")
	ar, err := grant(params)
	if err != nil {
		http.Redirect(w, r, errURL, 301)
		return
	}

	p, err := getProfile(ar.AccessToken)
	if err != nil {
		http.Redirect(w, r, errURL, 301)
		return
	}

	err = saveUser(ar, p)
	if err != nil {
		http.Redirect(w, r, errURL, 301)
		return
	}
	/*
		err = setCoupon(p.UserID)
		if err != nil {
			http.Redirect(w, r, errURL, 301)
			return
		}
	*/

	http.Redirect(w, r, sucURL, 301)
}

func showPage(w http.ResponseWriter, r *http.Request) {

	state := genState()

	u := url.Values{}
	u.Add("response_type", "code")
	u.Add("client_id", os.Getenv("WEB_CHANNEL_ID"))
	u.Add("redirect_uri", os.Getenv("WEB_REDIRECT_URL"))
	u.Add("state", state)

	url := fmt.Sprintf(urlFormat, u.Encode())

	surl := os.Getenv("WEB_STATIC_BASE_URL")

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	pageTempl.Execute(w, struct {
		LoginURL  string
		StaticURL string
	}{LoginURL: url, StaticURL: surl})
}

func showQR(w http.ResponseWriter, r *http.Request) {
	surl := os.Getenv("WEB_STATIC_BASE_URL")
	errURL := surl + "/error.html"

	params := r.URL.Query()
	if params.Get("t") == "" {
		http.Redirect(w, r, errURL, 301)
		return
	}

	curl := fmt.Sprintf("http://example.com/?hash=%s", params.Get("t"))

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	couponTempl.Execute(w, struct {
		StaticURL string
		CouponURL string
	}{StaticURL: surl, CouponURL: curl})
}

func createTable(svc *dynamodb.DynamoDB, w io.Writer) {
	// パラメータ
	tableInputParams := []*dynamodb.CreateTableInput{
		{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("user_id"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("user_id"),
					KeyType:       aws.String("HASH"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
			TableName: aws.String("users"),
		},
		{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("coupon_id"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("coupon_id"),
					KeyType:       aws.String("HASH"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
			TableName: aws.String("coupons"),
		},
	}

	for _, p := range tableInputParams {
		resp, err := svc.CreateTable(p)
		if err != nil {
			fmt.Fprintln(w, *p.TableName, err.Error())
			continue
		}

		fmt.Fprintln(w, *p.TableName, "created at", resp.TableDescription.CreationDateTime)
	}

}

func handleCreateTable(w http.ResponseWriter, r *http.Request) {
	svc, err := newService()
	if err != nil {
		fmt.Fprint(w, "error", err.Error())
	}
	createTable(svc, w)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", callback)
	mux.HandleFunc("/campaign", showPage)
	mux.HandleFunc("/coupon", showQR)
	//mux.HandleFunc("/create", handleCreateTable)

	ridge.Run(":8080", "", http.StripPrefix("/web", mux))
}
