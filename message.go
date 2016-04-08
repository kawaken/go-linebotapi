package linebotapi

import (
	"net/http"
)

const EndPointURL = "https://trialbot-api.line.me/v1/events"

var DefaultHeaders = map[string]interface{}{
	"Content-Type":                 "application/json; charset=UTF-8",
	"X-Line-ChannelID":             0,
	"X-Line-ChannelSecret":         "",
	"X-Line-Trusted-User-With-ACL": "",
}

const (
	ContentTypeText     = 1
	ContentTypeImage    = 2
	ContentTypeVideo    = 3
	ContentTypeAudio    = 4
	ContentTypeLocation = 7
	ContentTypeSticker  = 8
	ContentTypeContact  = 10
)

const (
	OpTypeAdded   = 4
	OpTypeBlocked = 8
)

type Location struct {
	Title    string  `json:"title"`
	Address  string  // this property is defined in reference, but sample does not exists.
	Latitude float64 `json:"latitude"`
}

type ContentMetadata struct {
	// Sticker
	PackageID string `json:"STKPKGID"`
	ID        string `json:"STKID"`
	Version   string `json:"STKVER"`
	Text      string `json:"STKTXT"`

	// Contact
	MID         string `json:"mid"`
	DisplayName string `json:"displayName"`
}

type OperationParams []string

func (o OperationParams) MID() string {
	if len(o) > 0 {
		return o[0]
	}
	return ""
}

type Content struct {
	ID          string   `json:"id"`
	ContentType int      `json:"contentType"`
	From        string   `json:"from"`
	CreatedTime int      `json:"createdTime"`
	To          []string `json:"to"`
	ToType      int      `json:"toType"`

	// Message
	ContentMetadata    *ContentMetadata `json:"contentMetadata"`
	Text               string           `json:"text"`
	OriginalContentURL string           `json:"originalContentUrl"`
	PreviewImageURL    string           `json:"previewImageUrl"`
	Location           *Location        `json:"location"`

	// Operation
	Revision int             `json:"revision"`
	OpType   int             `json:"opType"`
	Params   OperationParams `json:"params"`

	// MultipleMessage
	MessageNotified int        `json:"messageNotified"`
	Messages        []*Content `json:"messages"`
}

type Event struct {
	ID          string   `json:"id"`
	EventType   string   `json:"eventType"`
	From        string   `json:"from"`
	FromChannel string   `json:"fromChannel"`
	To          []string `json:"to"`
	ToChannel   int      `json:"toChannel"`
	Content     Content  `json:"content"`
}

type ReceivedData struct {
	Results []*Event `json:"result"`
}

func handler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":10080", nil)
}
