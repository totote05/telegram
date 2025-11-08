package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
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
}

func NewBot(token string) *Bot {
	return &Bot{
		token: token,
		client: &http.Client{
			Timeout: time.Second * 70, // un poco más que el timeout de long polling
		},
		offset:     0,
		apiBaseURL: apiURL, // Usar la constante por defecto
	}
}

func (b *Bot) makeRequest(ctx context.Context, method string, payload any) (*Response, error) {
	url := fmt.Sprintf(b.apiBaseURL, b.token, method)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error marshaling payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if !apiResp.Ok {
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

	log.Printf("Bot iniciado: @%s (%s)", user.Username, user.FirstName)
	return nil
}

func (b *Bot) handleMessage(ctx context.Context, msg *Message) {
	log.Printf("Mensaje de %s: %s", msg.From.FirstName, msg.Text)

	if strings.HasPrefix(msg.Text, "/") {
		if b.commandRegistry != nil {
			b.commandRegistry.Execute(ctx, b, msg)
		}
		return
	}

	if msg.Text != "" {
		response := fmt.Sprintf("Recibí tu mensaje: %s", msg.Text)
		if err := b.SendMessage(ctx, msg.Chat.ID, response); err != nil {
			log.Printf("Error enviando mensaje: %v", err)
		}
	}
}

func (b *Bot) SetCommandRegistry(commandRegistry *CommandRegistry) {
	b.commandRegistry = commandRegistry
}

func (b *Bot) Start(ctx context.Context) error {
	log.Println("Iniciando bot...")

	// Verificar que el token funciona
	if err := b.GetMe(ctx); err != nil {
		return fmt.Errorf("error verificando bot: %w", err)
	}

	log.Println("Esperando mensajes... (Ctrl+C para detener)")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown señalizado, cerrando bot...")
			return ctx.Err()
		default:
			updates, err := b.getUpdates(ctx)
			if err != nil {
				if ctx.Err() != nil {
					// El contexto fue cancelado, salir limpiamente
					return ctx.Err()
				}
				log.Printf("Error obteniendo updates: %v", err)
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
