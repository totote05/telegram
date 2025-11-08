package bot

import (
	"context"
	"testing"
)

func TestNewCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	if registry == nil {
		t.Error("expected non-nil CommandRegistry")
	}

	if registry.registry == nil {
		t.Error("expected registry map to be initialized")
	}

	if len(registry.registry) != 0 {
		t.Errorf("expected empty registry, got %d commands", len(registry.registry))
	}
}

func TestCommandRegistry_Register(t *testing.T) {
	registry := NewCommandRegistry()

	callCount := 0
	handler := func(ctx context.Context, bot *Bot, msg *Message) {
		callCount++
	}

	registry.Register("start", handler)

	if len(registry.registry) != 1 {
		t.Errorf("expected 1 command, got %d", len(registry.registry))
	}

	if registry.registry["start"] == nil {
		t.Error("expected handler to be registered")
	}
}

func TestCommandRegistry_Register_Multiple(t *testing.T) {
	registry := NewCommandRegistry()

	handler1 := func(ctx context.Context, bot *Bot, msg *Message) {}
	handler2 := func(ctx context.Context, bot *Bot, msg *Message) {}

	registry.Register("start", handler1)
	registry.Register("help", handler2)

	if len(registry.registry) != 2 {
		t.Errorf("expected 2 commands, got %d", len(registry.registry))
	}

	if registry.registry["start"] == nil {
		t.Error("expected start handler to be registered")
	}

	if registry.registry["help"] == nil {
		t.Error("expected help handler to be registered")
	}
}

func TestCommandRegistry_Register_Overwrite(t *testing.T) {
	registry := NewCommandRegistry()

	callCount1 := 0
	handler1 := func(ctx context.Context, bot *Bot, msg *Message) {
		callCount1++
	}

	callCount2 := 0
	handler2 := func(ctx context.Context, bot *Bot, msg *Message) {
		callCount2++
	}

	registry.Register("start", handler1)
	registry.Register("start", handler2) // Overwrite

	if len(registry.registry) != 1 {
		t.Errorf("expected 1 command, got %d", len(registry.registry))
	}

	// Verify the second handler is registered
	msg := &Message{Text: "/start", Chat: &Chat{ID: 123}}
	ctx := context.Background()
	bot := NewBot("test-token")
	registry.Execute(ctx, bot, msg)

	if callCount1 != 0 {
		t.Error("expected first handler not to be called")
	}

	if callCount2 != 1 {
		t.Error("expected second handler to be called")
	}
}

func TestCommandRegistry_Execute(t *testing.T) {
	tests := []struct {
		name          string
		msgText       string
		command       string
		handlerExists bool
		wantExecuted  bool
		wantCallCount int
	}{
		{
			name:          "execute existing command",
			msgText:       "/start",
			command:       "start",
			handlerExists: true,
			wantExecuted:  true,
			wantCallCount: 1,
		},
		{
			name:          "command with bot mention",
			msgText:       "/start@testbot",
			command:       "start",
			handlerExists: true,
			wantExecuted:  true,
			wantCallCount: 1,
		},
		{
			name:          "command with arguments",
			msgText:       "/start arg1 arg2",
			command:       "start",
			handlerExists: true,
			wantExecuted:  true,
			wantCallCount: 1,
		},
		{
			name:          "non-existent command",
			msgText:       "/unknown",
			command:       "start",
			handlerExists: true,
			wantExecuted:  false,
			wantCallCount: 0,
		},
		{
			name:          "message without slash prefix",
			msgText:       "start",
			command:       "start",
			handlerExists: true,
			wantExecuted:  false,
			wantCallCount: 0,
		},
		{
			name:          "empty message",
			msgText:       "",
			command:       "start",
			handlerExists: true,
			wantExecuted:  false,
			wantCallCount: 0,
		},
		{
			name:          "only slash",
			msgText:       "/",
			command:       "start",
			handlerExists: true,
			wantExecuted:  false,
			wantCallCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewCommandRegistry()

			callCount := 0
			var receivedMsg *Message
			var receivedBot *Bot

			if tt.handlerExists {
				handler := func(ctx context.Context, bot *Bot, msg *Message) {
					callCount++
					receivedMsg = msg
					receivedBot = bot
				}
				registry.Register(tt.command, handler)
			}

			msg := &Message{
				Text: tt.msgText,
				Chat: &Chat{ID: 123},
			}

			bot := NewBot("test-token")
			ctx := context.Background()

			executed := registry.Execute(ctx, bot, msg)

			if executed != tt.wantExecuted {
				t.Errorf("expected executed=%v, got %v", tt.wantExecuted, executed)
			}

			if callCount != tt.wantCallCount {
				t.Errorf("expected callCount=%d, got %d", tt.wantCallCount, callCount)
			}

			if tt.wantExecuted {
				if receivedMsg != msg {
					t.Error("expected handler to receive the message")
				}

				if receivedBot != bot {
					t.Error("expected handler to receive the bot")
				}
			}
		})
	}
}

func TestCommandRegistry_Execute_CommandExtraction(t *testing.T) {
	tests := []struct {
		name         string
		msgText      string
		expectedCmd  string
		shouldMatch  bool
	}{
		{
			name:        "simple command",
			msgText:     "/start",
			expectedCmd: "start",
			shouldMatch: true,
		},
		{
			name:        "command with bot mention",
			msgText:     "/start@mybot",
			expectedCmd: "start",
			shouldMatch: true,
		},
		{
			name:        "command with multiple parts",
			msgText:     "/help@mybot arg1 arg2",
			expectedCmd: "help",
			shouldMatch: true,
		},
		{
			name:        "different command",
			msgText:     "/help",
			expectedCmd: "start",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewCommandRegistry()

			callCount := 0
			handler := func(ctx context.Context, bot *Bot, msg *Message) {
				callCount++
			}

			registry.Register(tt.expectedCmd, handler)

			msg := &Message{
				Text: tt.msgText,
				Chat: &Chat{ID: 123},
			}

			bot := NewBot("test-token")
			ctx := context.Background()

			executed := registry.Execute(ctx, bot, msg)

			if executed != tt.shouldMatch {
				t.Errorf("expected executed=%v, got %v", tt.shouldMatch, executed)
			}

			expectedCalls := 0
			if tt.shouldMatch {
				expectedCalls = 1
			}

			if callCount != expectedCalls {
				t.Errorf("expected callCount=%d, got %d", expectedCalls, callCount)
			}
		})
	}
}

func TestCommandRegistry_Execute_Concurrent(t *testing.T) {
	registry := NewCommandRegistry()

	callCount := 0
	handler := func(ctx context.Context, bot *Bot, msg *Message) {
		callCount++
	}

	registry.Register("start", handler)

	bot := NewBot("test-token")
	ctx := context.Background()

	// Execute concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			msg := &Message{
				Text: "/start",
				Chat: &Chat{ID: 123},
			}
			registry.Execute(ctx, bot, msg)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if callCount != 10 {
		t.Errorf("expected callCount=10, got %d", callCount)
	}
}

