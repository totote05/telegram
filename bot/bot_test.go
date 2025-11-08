package bot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewBot(t *testing.T) {
	token := "test-token"
	bot := NewBot(token)

	if bot.token != token {
		t.Errorf("expected token %s, got %s", token, bot.token)
	}

	if bot.client == nil {
		t.Error("expected http client to be initialized")
	}

	if bot.offset != 0 {
		t.Errorf("expected offset 0, got %d", bot.offset)
	}

	if bot.client.Timeout != 70*time.Second {
		t.Errorf("expected timeout 70s, got %v", bot.client.Timeout)
	}
}

func TestBot_makeRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		payload        interface{}
		responseBody   string
		responseStatus int
		wantOk         bool
		wantErr        bool
		errContains    string
	}{
		{
			name:           "successful request",
			method:         "getMe",
			payload:        nil,
			responseBody:   `{"ok":true,"result":{"id":123,"first_name":"TestBot","username":"testbot"}}`,
			responseStatus: http.StatusOK,
			wantOk:         true,
			wantErr:        false,
		},
		{
			name:           "API error response",
			method:         "getMe",
			payload:        nil,
			responseBody:   `{"ok":false,"description":"Unauthorized"}`,
			responseStatus: http.StatusOK,
			wantOk:         false,
			wantErr:        true,
			errContains:    "API error",
		},
		{
			name:           "invalid JSON response",
			method:         "getMe",
			payload:        nil,
			responseBody:   `invalid json`,
			responseStatus: http.StatusOK,
			wantOk:         false,
			wantErr:        true,
			errContains:    "error unmarshaling response",
		},
		{
			name:           "request with payload",
			method:         "sendMessage",
			payload:        SendMessageRequest{ChatID: 123, Text: "Hello"},
			responseBody:   `{"ok":true,"result":{"message_id":1}}`,
			responseStatus: http.StatusOK,
			wantOk:         true,
			wantErr:        false,
		},
		{
			name:           "HTTP error",
			method:         "getMe",
			payload:        nil,
			responseBody:   ``,
			responseStatus: http.StatusInternalServerError,
			wantOk:         false,
			wantErr:        true,
			errContains:    "error", // Puede ser "error making request" o "error unmarshaling response"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Verificar que la URL contiene el método
				if !strings.Contains(r.URL.Path, tt.method) {
					t.Errorf("expected URL to contain method %s, got %s", tt.method, r.URL.Path)
				}

				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			bot := &Bot{
				token:     "test-token",
				client:    &http.Client{Timeout: 5 * time.Second},
				apiBaseURL: server.URL + "/bot%s/%s",
			}

			ctx := context.Background()
			resp, err := bot.makeRequest(ctx, tt.method, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("expected response, got nil")
				return
			}

			if resp.Ok != tt.wantOk {
				t.Errorf("expected Ok=%v, got %v", tt.wantOk, resp.Ok)
			}
		})
	}
}

func TestBot_makeRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	bot := &Bot{
		token:     "test-token",
		client:    &http.Client{Timeout: 5 * time.Second},
		apiBaseURL: server.URL + "/bot%s/%s",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := bot.makeRequest(ctx, "getMe", nil)
	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}
}

func TestBot_getUpdates(t *testing.T) {
	tests := []struct {
		name        string
		offset      int
		response    string
		wantUpdates int
		wantErr     bool
	}{
		{
			name:        "successful getUpdates",
			offset:      0,
			response:    `[{"update_id":1,"message":{"message_id":1,"chat":{"id":123,"type":"private"},"text":"Hello"}}]`,
			wantUpdates: 1,
			wantErr:     false,
		},
		{
			name:        "empty updates",
			offset:      0,
			response:    `[]`,
			wantUpdates: 0,
			wantErr:     false,
		},
		{
			name:        "invalid updates JSON",
			offset:      0,
			response:    `{"ok":true,"result":"invalid"}`,
			wantUpdates: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ok":true,"result":` + tt.response + `}`))
			}))
			defer server.Close()

			bot := &Bot{
				token:     "test-token",
				client:    &http.Client{Timeout: 5 * time.Second},
				offset:    tt.offset,
				apiBaseURL: server.URL + "/bot%s/%s",
			}

			ctx := context.Background()
			updates, err := bot.getUpdates(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(updates) != tt.wantUpdates {
				t.Errorf("expected %d updates, got %d", tt.wantUpdates, len(updates))
			}
		})
	}
}

func TestBot_SendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("error decoding request: %v", err)
		}

		if req.ChatID != 123 {
			t.Errorf("expected ChatID 123, got %d", req.ChatID)
		}

		if req.Text != "Hello" {
			t.Errorf("expected Text 'Hello', got %s", req.Text)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer server.Close()

	bot := &Bot{
		token:     "test-token",
		client:    &http.Client{Timeout: 5 * time.Second},
		apiBaseURL: server.URL + "/bot%s/%s",
	}

	ctx := context.Background()
	err := bot.SendMessage(ctx, 123, "Hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBot_SendMessage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":false,"description":"Bad Request"}`))
	}))
	defer server.Close()

	bot := &Bot{
		token:     "test-token",
		client:    &http.Client{Timeout: 5 * time.Second},
		apiBaseURL: server.URL + "/bot%s/%s",
	}

	ctx := context.Background()
	err := bot.SendMessage(ctx, 123, "Hello")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestBot_GetMe(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful GetMe",
			response:    `{"ok":true,"result":{"id":123,"first_name":"TestBot","username":"testbot"}}`,
			wantErr:     false,
		},
		{
			name:        "API error",
			response:    `{"ok":false,"description":"Unauthorized"}`,
			wantErr:     true,
			errContains: "API error",
		},
		{
			name:        "invalid user JSON",
			response:    `{"ok":true,"result":"invalid"}`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			bot := &Bot{
				token:     "test-token",
				client:    &http.Client{Timeout: 5 * time.Second},
				apiBaseURL: server.URL + "/bot%s/%s",
			}

			ctx := context.Background()
			err := bot.GetMe(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBot_handleMessage(t *testing.T) {
	tests := []struct {
		name           string
		msg            *Message
		hasRegistry    bool
		commandHandler func(context.Context, *Bot, *Message)
		wantCallCount  int
	}{
		{
			name: "message with command and registry",
			msg: &Message{
				Text: "/start",
				From: &User{FirstName: "Test"},
				Chat: &Chat{ID: 123},
			},
			hasRegistry: true,
			commandHandler: func(ctx context.Context, b *Bot, m *Message) {
				// Mock handler
			},
			wantCallCount: 1,
		},
		{
			name: "regular message",
			msg: &Message{
				Text: "Hello",
				From: &User{FirstName: "Test"},
				Chat: &Chat{ID: 123},
			},
			hasRegistry: false,
			wantCallCount: 0,
		},
		{
			name: "empty message",
			msg: &Message{
				Text: "",
				From: &User{FirstName: "Test"},
				Chat: &Chat{ID: 123},
			},
			hasRegistry: false,
			wantCallCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			var sentMessage string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req SendMessageRequest
				json.NewDecoder(r.Body).Decode(&req)
				sentMessage = req.Text
				callCount++
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
			}))
			defer server.Close()

			bot := &Bot{
				token:     "test-token",
				client:    &http.Client{Timeout: 5 * time.Second},
				apiBaseURL: server.URL + "/bot%s/%s",
			}

			if tt.hasRegistry {
				registry := NewCommandRegistry()
				if tt.commandHandler != nil {
					registry.Register("start", tt.commandHandler)
				}
				bot.SetCommandRegistry(registry)
			}

			ctx := context.Background()
			bot.handleMessage(ctx, tt.msg)

			// Give goroutines time to complete
			time.Sleep(50 * time.Millisecond)

			if !strings.HasPrefix(tt.msg.Text, "/") && tt.msg.Text != "" {
				if !strings.Contains(sentMessage, tt.msg.Text) {
					t.Errorf("expected message to contain %q, got %q", tt.msg.Text, sentMessage)
				}
			}
		})
	}
}

func TestBot_handleMessage_NilRegistry(t *testing.T) {
	// Test that handleMessage doesn't panic when registry is nil
	msg := &Message{
		Text: "/start",
		From: &User{FirstName: "Test"},
		Chat: &Chat{ID: 123},
	}

	bot := &Bot{
		token:     "test-token",
		client:    &http.Client{Timeout: 5 * time.Second},
		apiBaseURL: "https://api.telegram.org/bot%s/%s",
		// commandRegistry is nil
	}

	ctx := context.Background()
	// This should not panic
	bot.handleMessage(ctx, msg)
}

func TestBot_SetCommandRegistry(t *testing.T) {
	bot := NewBot("test-token")
	registry := NewCommandRegistry()

	bot.SetCommandRegistry(registry)

	if bot.commandRegistry != registry {
		t.Error("expected command registry to be set")
	}
}

func TestBot_Start_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":[]}`))
	}))
	defer server.Close()

	bot := &Bot{
		token:     "test-token",
		client:    &http.Client{Timeout: 5 * time.Second},
		apiBaseURL: server.URL + "/bot%s/%s",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := bot.Start(ctx)
	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}
	// El error puede ser context.Canceled o un error envuelto
	if err != context.Canceled && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context.Canceled or wrapped error, got %v", err)
	}
}

func TestBot_Start_GetMeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":false,"description":"Unauthorized"}`))
	}))
	defer server.Close()

	bot := &Bot{
		token:     "test-token",
		client:    &http.Client{Timeout: 5 * time.Second},
		apiBaseURL: server.URL + "/bot%s/%s",
	}

	ctx := context.Background()
	err := bot.Start(ctx)
	if err == nil {
		t.Error("expected error from GetMe, got nil")
	}
	if !strings.Contains(err.Error(), "error verificando bot") {
		t.Errorf("expected error about bot verification, got %v", err)
	}
}

