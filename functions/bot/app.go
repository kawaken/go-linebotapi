package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fujiwara/ridge"
	"github.com/kawaken/rod"
	"github.com/line/line-bot-sdk-go/linebot"
)

type user struct {
	UserID       string
	DisplayName  string
	RegisteredAt string
	Hash         string
}

func getUser(userID string) (*user, error) {
	svc, err := newService()
	if err != nil {
		return nil, err
	}

	params := &dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
	}

	// GetItemの実行
	resp, err := svc.GetItem(params)
	if err != nil {
		return nil, err
	}

	u := &user{
		UserID: userID,
	}

	if len(resp.Item) > 0 {
		if dn, ok := resp.Item["display_name"]; ok {
			u.DisplayName = *dn.S
		}
		if ra, ok := resp.Item["registered_at"]; ok {
			u.RegisteredAt = *ra.N
		}
		if h, ok := resp.Item["hash"]; ok {
			u.Hash = *h.S
		}
	}

	return u, nil
}

/*
func hasCoupons(userID string) (bool, error) {
	svc, err := newService()
	if err != nil {
		return false, err
	}

	params := &dynamodb.GetItemInput{
		TableName: aws.String("coupons"),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
	}

	// GetItemの実行
	resp, err := svc.GetItem(params)
	if err != nil {
		return false, err
	}

	return (len(resp.Item) > 0), nil
}
*/

func main() {
	logger := log.New(os.Stderr, "", log.Lshortfile)

	handler, err := rod.New(os.Getenv("BOT_CHANNEL_SECRET"), os.Getenv("BOT_CHANNEL_TOKEN"))
	if err != nil {
		logger.Print(err)
		return
	}

	handler.OnTextMessageRecieved = func(e *linebot.Event, m *linebot.TextMessage) []linebot.Message {

		var messages []linebot.Message

		switch m.Text {
		case "クーポン":
			if e.Source.Type != linebot.EventSourceTypeUser {
				return nil
			}

			user, err := getUser(e.Source.UserID)
			if err != nil {
				log.Print(err.Error())
				return nil
			}

			if user.RegisteredAt != "" {
				t := os.Getenv("WEB_STATIC_BASE_URL") + "/yakiniku2.jpg"
				q := fmt.Sprintf("%s?t=%s", os.Getenv("WEB_COUPON_URL"), user.Hash)
				messages = append(messages,
					linebot.NewTemplateMessage("クーポン",
						linebot.NewButtonsTemplate(t, "50%オフクーポン", "お会計時に料金の50%を割引いたします。",
							linebot.NewURITemplateAction("使用する(QRコードを表示)", q),
						),
					),
				)
			} else {
				messages = append(messages,
					linebot.NewTextMessage("クーポンがありません\nキャンペーンに応募してゲットしよう!!!"),
					linebot.NewTextMessage(os.Getenv("WEB_CAMPAIGN_URL")),
				)
			}
		case "キャンペーン":
			if e.Source.Type != linebot.EventSourceTypeUser {
				return nil
			}
			messages = append(messages,
				linebot.NewTextMessage("クーポンをゲットしよう!!!"),
				linebot.NewTextMessage(os.Getenv("WEB_CAMPAIGN_URL")),
			)
		}

		return messages
	}

	handler.OnFollowed = func(e *linebot.Event, userID string) []linebot.Message {
		user, err := getUser(userID)
		if err != nil {
			log.Print(err.Error())
			return nil
		}

		messages := []linebot.Message{
			linebot.NewTextMessage("友達登録ありがとうございます。"),
		}

		if user.RegisteredAt != "" {
			tm := `キャンペーンのご応募ありがとうございます！
「クーポン」とメッセージを送っていただくとクーポンが表示されます。`

			messages = append(messages,
				linebot.NewTextMessage(tm),
				linebot.NewStickerMessage("2", "41"),
			)

			return messages
		}

		messages = append(messages,
			linebot.NewTextMessage("キャンペーンに応募してクーポンをGETしよう！"),
			linebot.NewStickerMessage("1", "114"),
			linebot.NewTextMessage(os.Getenv("WEB_CAMPAIGN_URL")),
		)

		return messages
	}

	ridge.Run(":8080", "", http.StripPrefix("/bot/callback", handler))
}
