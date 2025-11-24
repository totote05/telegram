package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/totote05/telegram/bot"
)

func commandStart(ctx context.Context, b *bot.Bot, msg *bot.Message) {
	welcome := "¡Hola! Soy un bot de Telegram."
	if err := b.SendMessage(ctx, msg.Chat.ID, welcome); err != nil {
		log.Printf("Error enviando bienvenida: %v", err)
	}
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
	}

	commands := bot.NewCommandRegistry()
	commands.Register("start", commandStart)
	b := bot.NewBot(token, bot.WithCommandRegistry(commands))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Señal recibida, cerrando...")
		cancel()
	}()

	if err := b.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Error ejecutando bot: %v", err)
	}
}
