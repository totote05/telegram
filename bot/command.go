package bot

import (
	"context"
	"strings"
)

type (
	CommandRegistry struct {
		registry map[string]Command
	}
	Command func(context.Context, *Bot, *Message)
)

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		registry: make(map[string]Command),
	}
}

func (cr *CommandRegistry) Register(command string, action Command) {
	cr.registry[command] = action
}

func (cr *CommandRegistry) Execute(ctx context.Context, bot *Bot, msg *Message) bool {
	if !strings.HasPrefix(msg.Text, "/") {
		return false
	}

	parts := strings.Fields(msg.Text)
	if len(parts) == 0 {
		return false
	}

	command := strings.TrimPrefix(parts[0], "/")
	// Remover @botname si está presente
	command = strings.Split(command, "@")[0]

	action, exists := cr.registry[command]
	if !exists {
		return false
	}

	action(ctx, bot, msg)
	return true
}
