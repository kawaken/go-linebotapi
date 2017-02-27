package main

import (
	"log"
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
				N: aws.String(userID),
			},
		},
	}

	// GetItemの実行
	resp, err := svc.GetItem(params)
	if err != nil {
		return nil, err
	}

	u := &user{
		UserID:       userID,
		DisplayName:  *resp.Item["display_name"].S,
		RegisteredAt: *resp.Item["registered_at"].N,
	}

	return u, nil
}

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
			messages = append(messages, &linebot.TextMessage{Text: "クーポンがありません"})
		}

		return messages
	}

	handler.OnFollowed = func(e *linebot.Event, userID string) []linebot.Message {
		user, err := getUser(userID)
		if err != nil {
			log.Print(err.Error())
		}

		messages := []linebot.Message{
			linebot.NewTextMessage("友達登録ありがとうございます。"),
		}

		if user.RegisteredAt != "" {
			tm := `キャンペーンのご応募もありがとうございます！
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

	ridge.Run(":8080", "/callback", handler)
}
