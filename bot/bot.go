package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/totote05/go-toolkit/pkg/logger"
)

const (
	apiURL  = "https://api.telegram.org/bot%s/%s"
	timeout = 60 // segundos para long polling
)

type Bot struct {
	token           string
	client          *http.Client
	offset          int
	commandRegistry *CommandRegistry
	apiBaseURL      string // Para testing, por defecto usa la constante apiURL
	logger          *slog.Logger
}

// BotOption es una función que configura opciones del Bot.
type BotOption func(*Bot)

// WithLogger configura el logger que utilizará el bot.
// Permite que el consumidor reutilice su propia instancia de logger.
//
// Ejemplo:
//
//	appLogger := slog.New(customHandler)
//	bot := bot.NewBot(token, bot.WithLogger(appLogger))
func WithLogger(log *slog.Logger) BotOption {
	return func(b *Bot) {
		if log != nil {
			b.logger = log
		}
	}
}

// WithCommandRegistry configura el registro de comandos del bot.
//
// Ejemplo:
//
//	commands := bot.NewCommandRegistry()
//	commands.Register("start", commandStart)
//	bot := bot.NewBot(token, bot.WithCommandRegistry(commands))
func WithCommandRegistry(registry *CommandRegistry) BotOption {
	return func(b *Bot) {
		b.commandRegistry = registry
	}
}

// defaultLogger crea un logger por defecto usando el handler de go-toolkit.
func defaultLogger() *slog.Logger {
	handler := logger.NewHandler(os.Stdout, &logger.HandlerOptions{
		Level:      slog.LevelInfo,
		StackTrace: 5,
	})
	return slog.New(handler)
}

// NewBot crea una nueva instancia del bot con el token proporcionado.
// Acepta opciones funcionales para configurar el bot.
//
// Ejemplo de uso básico:
//
//	bot := bot.NewBot(token)
//
// Ejemplo con logger personalizado:
//
//	appLogger := slog.New(customHandler)
//	bot := bot.NewBot(token, bot.WithLogger(appLogger))
//
// Ejemplo con logger y comandos:
//
//	bot := bot.NewBot(token,
//	    bot.WithLogger(appLogger),
//	    bot.WithCommandRegistry(commands),
//	)
func NewBot(token string, opts ...BotOption) *Bot {
	b := &Bot{
		token: token,
		client: &http.Client{
			Timeout: time.Second * 70, // un poco más que el timeout de long polling
		},
		offset:     0,
		apiBaseURL: apiURL,          // Usar la constante por defecto
		logger:     defaultLogger(), // Logger por defecto
	}

	// Aplicar opciones
	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *Bot) makeRequest(ctx context.Context, method string, payload any) (*Response, error) {
	url := fmt.Sprintf(b.apiBaseURL, b.token, method)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			b.logger.Error("Error marshaling payload",
				slog.String("method", method),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("error marshaling payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		b.logger.Error("Error creating request",
			slog.String("method", method),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	b.logger.Debug("Enviando request a Telegram API",
		slog.String("method", method),
	)

	resp, err := b.client.Do(req)
	if err != nil {
		b.logger.Error("Error making request",
			slog.String("method", method),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("Error reading response",
			slog.String("method", method),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		b.logger.Error("Error unmarshaling response",
			slog.String("method", method),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if !apiResp.Ok {
		b.logger.Error("API error response",
			slog.String("method", method),
			slog.String("description", apiResp.Description),
		)
		return nil, fmt.Errorf("API error: %s", apiResp.Description)
	}

	return &apiResp, nil
}

func (b *Bot) getUpdates(ctx context.Context) ([]Update, error) {
	params := map[string]interface{}{
		"offset":  b.offset,
		"timeout": timeout,
	}

	resp, err := b.makeRequest(ctx, "getUpdates", params)
	if err != nil {
		return nil, err
	}

	var updates []Update
	if err := json.Unmarshal(resp.Result, &updates); err != nil {
		return nil, fmt.Errorf("error unmarshaling updates: %w", err)
	}

	return updates, nil
}

func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	payload := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	_, err := b.makeRequest(ctx, "sendMessage", payload)
	return err
}

func (b *Bot) GetMe(ctx context.Context) error {
	resp, err := b.makeRequest(ctx, "getMe", nil)
	if err != nil {
		return err
	}

	var user User
	if err := json.Unmarshal(resp.Result, &user); err != nil {
		return err
	}

	b.logger.Info("Bot iniciado",
		slog.String("username", user.Username),
		slog.String("first_name", user.FirstName),
	)
	return nil
}

func (b *Bot) handleMessage(ctx context.Context, msg *Message) {
	b.logger.Info("Mensaje recibido",
		slog.String("from", msg.From.FirstName),
		slog.Int64("chat_id", msg.Chat.ID),
		slog.String("text", msg.Text),
	)

	if strings.HasPrefix(msg.Text, "/") {
		if b.commandRegistry != nil {
			b.commandRegistry.Execute(ctx, b, msg)
		}
		return
	}

	if msg.Text != "" {
		response := fmt.Sprintf("Recibí tu mensaje: %s", msg.Text)
		if err := b.SendMessage(ctx, msg.Chat.ID, response); err != nil {
			b.logger.Error("Error enviando mensaje",
				slog.Int64("chat_id", msg.Chat.ID),
				slog.String("error", err.Error()),
			)
		}
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Iniciando bot...")

	// Verificar que el token funciona
	if err := b.GetMe(ctx); err != nil {
		return fmt.Errorf("error verificando bot: %w", err)
	}

	b.logger.Info("Esperando mensajes... (Ctrl+C para detener)")

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Shutdown señalizado, cerrando bot...")
			return ctx.Err()
		default:
			updates, err := b.getUpdates(ctx)
			if err != nil {
				if ctx.Err() != nil {
					// El contexto fue cancelado, salir limpiamente
					return ctx.Err()
				}
				b.logger.Error("Error obteniendo updates",
					slog.String("error", err.Error()),
				)
				time.Sleep(3 * time.Second)
				continue
			}

			for _, update := range updates {
				// Actualizar offset para el próximo request
				b.offset = update.UpdateID + 1

				// Procesar mensaje en goroutine para no bloquear
				if update.Message != nil {
					msg := update.Message
					go b.handleMessage(ctx, msg)
				}
			}
		}
	}
}
