# Guía de Inicio Rápido

Esta guía te ayudará a crear tu primer bot de Telegram usando esta librería.

## Paso 1: Obtener un Token de Bot

1. Abre Telegram y busca [@BotFather](https://t.me/BotFather)
2. Envía el comando `/newbot`
3. Sigue las instrucciones para crear tu bot
4. Copia el token que te proporciona BotFather

## Paso 2: Configurar el Proyecto

Crea un nuevo directorio para tu proyecto:

```bash
mkdir my-telegram-bot
cd my-telegram-bot
go mod init my-telegram-bot
```

## Paso 3: Instalar la Dependencia

```bash
go get github.com/totote05/telegram
```

## Paso 4: Crear el Bot Básico

Crea un archivo `main.go`:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/totote05/telegram/bot"
)

func main() {
    // Obtener el token desde una variable de entorno
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    // Crear una nueva instancia del bot
    b := bot.NewBot(token)

    // Configurar el contexto con cancelación para shutdown graceful
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Manejar señales del sistema (Ctrl+C, SIGTERM)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Señal recibida, cerrando...")
        cancel()
    }()

    // Iniciar el bot
    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error ejecutando bot: %v", err)
    }
}
```

## Paso 5: Ejecutar el Bot

```bash
export TELEGRAM_BOT_TOKEN="tu-token-aqui"
go run main.go
```

## Paso 6: Agregar Comandos

Para agregar comandos personalizados, usa el `CommandRegistry`:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/totote05/telegram/bot"
)

// Handler para el comando /start
func commandStart(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    welcome := "¡Hola! Soy un bot de Telegram. Usa /help para ver los comandos disponibles."
    if err := b.SendMessage(ctx, msg.Chat.ID, welcome); err != nil {
        log.Printf("Error enviando mensaje: %v", err)
    }
}

// Handler para el comando /help
func commandHelp(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    helpText := "Comandos disponibles:\n/start - Iniciar el bot\n/help - Mostrar esta ayuda"
    if err := b.SendMessage(ctx, msg.Chat.ID, helpText); err != nil {
        log.Printf("Error enviando mensaje: %v", err)
    }
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    // Crear y configurar el registro de comandos
    commands := bot.NewCommandRegistry()
    commands.Register("start", commandStart)
    commands.Register("help", commandHelp)
    
    // Crear el bot con el registro de comandos
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
```

## Próximos Pasos

- Lee la [Referencia de API](./api-reference.md) para conocer todas las funciones disponibles
- Consulta la [Guía de Comandos](./commands.md) para aprender a crear comandos más complejos
- Revisa los [Ejemplos](./examples.md) para ver casos de uso avanzados

