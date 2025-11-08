# Ejemplos

Esta página contiene ejemplos prácticos de uso de la librería.

## Ejemplo Básico

El ejemplo más simple de un bot:

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
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Bot con Comandos

Bot con varios comandos registrados:

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

func commandStart(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    welcome := "¡Hola! Soy un bot de Telegram."
    if err := b.SendMessage(ctx, msg.Chat.ID, welcome); err != nil {
        log.Printf("Error: %v", err)
    }
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)
    commands := bot.NewCommandRegistry()
    commands.Register("start", commandStart)
    b.SetCommandRegistry(commands)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Bot con Múltiples Comandos

Ejemplo con varios comandos y manejo de argumentos:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/totote05/telegram/bot"
)

func commandStart(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    b.SendMessage(ctx, msg.Chat.ID, "¡Bienvenido! Usa /help para ver los comandos.")
}

func commandHelp(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    help := `Comandos disponibles:
/start - Iniciar
/help - Ayuda
/time - Hora actual
/echo <texto> - Repetir texto`
    b.SendMessage(ctx, msg.Chat.ID, help)
}

func commandTime(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    now := time.Now().Format("2006-01-02 15:04:05")
    b.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Hora: %s", now))
}

func commandEcho(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    parts := strings.Fields(msg.Text)
    if len(parts) < 2 {
        b.SendMessage(ctx, msg.Chat.ID, "Uso: /echo <texto>")
        return
    }
    text := strings.Join(parts[1:], " ")
    b.SendMessage(ctx, msg.Chat.ID, text)
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)
    commands := bot.NewCommandRegistry()
    commands.Register("start", commandStart)
    commands.Register("help", commandHelp)
    commands.Register("time", commandTime)
    commands.Register("echo", commandEcho)
    b.SetCommandRegistry(commands)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Bot con Timeout

Ejemplo usando context con timeout para operaciones:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/totote05/telegram/bot"
)

func commandSlow(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    // Crear contexto con timeout para esta operación
    opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Simular operación lenta
    select {
    case <-time.After(3 * time.Second):
        b.SendMessage(ctx, msg.Chat.ID, "Operación completada")
    case <-opCtx.Done():
        b.SendMessage(ctx, msg.Chat.ID, "Operación cancelada por timeout")
    }
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)
    commands := bot.NewCommandRegistry()
    commands.Register("slow", commandSlow)
    b.SetCommandRegistry(commands)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Bot con Estado Compartido

Ejemplo usando mutex para compartir estado entre comandos:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "sync"
    "syscall"

    "github.com/totote05/telegram/bot"
)

type Counter struct {
    mu    sync.RWMutex
    count int
}

var counter = &Counter{count: 0}

func commandIncrement(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    counter.mu.Lock()
    counter.count++
    count := counter.count
    counter.mu.Unlock()

    b.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Contador: %d", count))
}

func commandGet(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    counter.mu.RLock()
    count := counter.count
    counter.mu.RUnlock()

    b.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Contador actual: %d", count))
}

func commandSet(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    parts := strings.Fields(msg.Text)
    if len(parts) < 2 {
        b.SendMessage(ctx, msg.Chat.ID, "Uso: /set <número>")
        return
    }

    value, err := strconv.Atoi(parts[1])
    if err != nil {
        b.SendMessage(ctx, msg.Chat.ID, "Error: número inválido")
        return
    }

    counter.mu.Lock()
    counter.count = value
    counter.mu.Unlock()

    b.SendMessage(ctx, msg.Chat.ID, fmt.Sprintf("Contador establecido a: %d", value))
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)
    commands := bot.NewCommandRegistry()
    commands.Register("inc", commandIncrement)
    commands.Register("get", commandGet)
    commands.Register("set", commandSet)
    b.SetCommandRegistry(commands)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Bot con Validación de Usuario

Ejemplo que valida si el usuario tiene permisos:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/totote05/telegram/bot"
)

// Lista de IDs de usuarios permitidos
var allowedUsers = map[int64]bool{
    123456789: true, // Reemplaza con tu ID
}

func isAllowed(userID int64) bool {
    return allowedUsers[userID]
}

func commandAdmin(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    if !isAllowed(msg.From.ID) {
        b.SendMessage(ctx, msg.Chat.ID, "No tienes permisos para usar este comando")
        return
    }

    b.SendMessage(ctx, msg.Chat.ID, "Comando de administrador ejecutado")
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)
    commands := bot.NewCommandRegistry()
    commands.Register("admin", commandAdmin)
    b.SetCommandRegistry(commands)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Bot con Manejo de Errores Robusto

Ejemplo con manejo de errores completo:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/totote05/telegram/bot"
)

func sendMessageWithRetry(ctx context.Context, b *bot.Bot, chatID int64, text string, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := b.SendMessage(ctx, chatID, text)
        if err == nil {
            return nil
        }
        
        lastErr = err
        log.Printf("Intento %d fallido: %v", i+1, err)
        
        // Esperar antes de reintentar
        select {
        case <-time.After(time.Second * time.Duration(i+1)):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return fmt.Errorf("falló después de %d intentos: %w", maxRetries, lastErr)
}

func commandReliable(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    err := sendMessageWithRetry(ctx, b, msg.Chat.ID, "Mensaje confiable", 3)
    if err != nil {
        log.Printf("Error crítico enviando mensaje: %v", err)
    }
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN no está definido")
    }

    b := bot.NewBot(token)
    commands := bot.NewCommandRegistry()
    commands.Register("reliable", commandReliable)
    b.SetCommandRegistry(commands)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error: %v", err)
    }
}
```

## Ver Más

- Consulta la [Guía de Inicio Rápido](./getting-started.md) para comenzar
- Lee la [Referencia de API](./api-reference.md) para detalles completos
- Revisa la [Guía de Comandos](./commands.md) para aprender más sobre comandos

