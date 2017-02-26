package main

import (
	"log"
	"os"

	"github.com/fujiwara/ridge"
	"github.com/kawaken/rod"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	logger := log.New(os.Stderr, "", log.Lshortfile)

	handler, err := rod.New(os.Getenv("BOT_CHANNEL_SECRET"), os.Getenv("BOT_CHANNEL_TOKEN"))
	if err != nil {
		logger.Print(err)
		return
	}

	handler.OnTextMessageRecieved = func(e *linebot.Event, m *linebot.TextMessage) []linebot.Message {
		return []linebot.Message{
			&linebot.TextMessage{Text: m.Text},
		}
	}

	ridge.Run(":8080", "/callback", handler)
}
