package bot

import (
	"encoding/json"
)

type (
	Update struct {
		UpdateID int      `json:"update_id"`
		Message  *Message `json:"message,omitempty"`
	}

	Message struct {
		MessageID int    `json:"message_id"`
		From      *User  `json:"from,omitempty"`
		Chat      *Chat  `json:"chat"`
		Date      int64  `json:"date"`
		Text      string `json:"text,omitempty"`
	}

	User struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name,omitempty"`
		Username  string `json:"username,omitempty"`
	}

	Chat struct {
		ID       int64  `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title,omitempty"`
		Username string `json:"username,omitempty"`
	}

	Response struct {
		Ok          bool            `json:"ok"`
		Result      json.RawMessage `json:"result,omitempty"`
		Description string          `json:"description,omitempty"`
	}

	SendMessageRequest struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}
)
