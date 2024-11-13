package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type MessageResponse struct {
	MessageID string  `json:"message_id"`
	RootID    string  `json:"root_id"`
	ParentID  string  `json:"parent_id"`
	ThreadID  string  `json:"thread_id"`
	MsgType   string  `json:"msg_type"`
	Content   Content `json:"content"`
}

// Function to map ReplyMessageResp to your custom MessageResponse
func mapToMessageResponse(replyResp *larkim.ReplyMessageResp) (*MessageResponse, error) {
	if replyResp == nil || replyResp.Data == nil {
		return nil, fmt.Errorf("invalid response: data is nil")
	}

	// var sender Sender
	// idType := getStringValue(replyResp.Data.Sender.IdType)
	// switch idType {
	// case "open_id":
	// 	sender.OpenID = getStringValue(replyResp.Data.Sender.Id)
	// default: // No matching ID type
	// 	return nil, fmt.Errorf("invalid sender ID type: %s", idType)
	// }

	messageResponse := &MessageResponse{
		MessageID: getStringValue(replyResp.Data.MessageId),
		RootID:    getStringValue(replyResp.Data.RootId),
		ParentID:  getStringValue(replyResp.Data.ParentId),
		ThreadID:  getStringValue(replyResp.Data.ThreadId),
		MsgType:   getStringValue(replyResp.Data.MsgType),
	}

	// Handle the message body if it is present
	if replyResp.Data.Body != nil && replyResp.Data.Body.Content != nil {
		var content Content
		content.Text = getStringValue(replyResp.Data.Body.Content)
		messageResponse.Content = content
	}

	return messageResponse, nil
}

// Helper function to get string value from a pointer
func getStringValue(strPtr *string) string {
	if strPtr != nil {
		return *strPtr
	}
	return ""
}

type EventBody struct {
	Event Event `json:"event"`
}

type Event struct {
	Sender  Sender  `json:"sender"`
	Message Message `json:"message"`
}

type Sender struct {
	UnionID string `json:"union_id"`
	UserID  string `json:"user_id"`
	OpenID  string `json:"open_id"`
}

func (s *Sender) UnmarshalJSON(data []byte) error {
	type Alias struct {
		SenderID struct {
			UnionId string `json:"union_id"`
			UserId  string `json:"user_id"`
			OpenId  string `json:"open_id"`
		} `json:"sender_id"`
	}

	var a Alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	s.UnionID = a.SenderID.UnionId
	s.UserID = a.SenderID.UserId
	s.OpenID = a.SenderID.OpenId

	return nil
}

func (s Sender) MarshalJSON() ([]byte, error) {
	type Alias struct {
		SenderID struct {
			UnionId string `json:"union_id"`
			UserId  string `json:"user_id"`
			OpenId  string `json:"open_id"`
		} `json:"sender_id"`
	}

	var a Alias
	a.SenderID.UnionId = s.UnionID
	a.SenderID.UserId = s.UserID
	a.SenderID.OpenId = s.OpenID

	return json.Marshal(a)
}

type Message struct {
	MessageID string   `json:"message_id"`
	RootID    string   `json:"root_id"`
	ParentID  string   `json:"parent_id"`
	ThreadID  string   `json:"thread_id"`
	Content   Content  `json:"content"`
	Mentions  []string `json:"mentions"`
	// etc.
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias struct {
		MessageID string   `json:"message_id"`
		RootID    string   `json:"root_id"`
		ParentID  string   `json:"parent_id"`
		ThreadID  string   `json:"thread_id"`
		Content   string   `json:"content"`
		Mentions  []struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"mentions"`
	}

	var a Alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.MessageID = a.MessageID
	m.RootID = a.RootID
	m.ParentID = a.ParentID
	m.ThreadID = a.ThreadID

	if err := json.Unmarshal([]byte(a.Content), &m.Content); err != nil {
		return fmt.Errorf("Error unmarshalling Content field: %v", err)
	}

	// Remove mention keys from content text
	if a.Mentions != nil {
		var mentionKeys []string
		for _, mention := range a.Mentions {
			mentionKeys = append(mentionKeys, mention.Key)
			m.Mentions = append(m.Mentions, mention.Name)
		}
		// Remove mention keys from the content text
		cleanedText := removeMentions(m.Content.Text, mentionKeys)
		m.Content.Text = cleanedText
	}

	return nil
}

func (m Message) MarshalJSON() ([]byte, error) {
	type Alias struct {
		MessageID string `json:"message_id"`
		RootID    string `json:"root_id"`
		ParentID  string `json:"parent_id"`
		ThreadID  string `json:"thread_id"`
		Content   string `json:"content"`
	}

	var a Alias
	a.MessageID = m.MessageID
	a.RootID = m.RootID
	a.ParentID = m.ParentID
	a.ThreadID = m.ThreadID

	contentBytes, err := json.Marshal(m.Content)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling Content field: %v", err)
	}
	a.Content = string(contentBytes)

	return json.Marshal(a)
}

func (m *Message) ContainesBotMention() bool {
	for _, m := range m.Mentions {
		if m == "VM-Manager" {
			return true
		}
	}
	return false
}

func (m *Message) ContainsMention(mention string) bool {
	for _, m := range m.Mentions {
		if m == mention {
			return true
		}
	}
	return false
}

type Content struct {
	Text string `json:"text"`
}

// Helper function to remove mention keys from content text
func removeMentions(text string, mentionKeys []string) string {
	cleanedText := text
	for _, key := range mentionKeys {
		// Use regex to remove mention key along with any surrounding whitespace
		regex := regexp.MustCompile(fmt.Sprintf(`\s*%s\s*`, regexp.QuoteMeta(key)))
		cleanedText = regex.ReplaceAllString(cleanedText, " ")
	}

	// Trim any leading/trailing whitespace after removal
	return strings.TrimSpace(cleanedText)
}
