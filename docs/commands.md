# Guía de Comandos

Esta guía explica cómo crear, registrar y gestionar comandos en tu bot de Telegram.

## Conceptos Básicos

Un comando en Telegram es un mensaje que comienza con `/` seguido del nombre del comando. Por ejemplo:
- `/start` - Comando start
- `/help` - Comando help
- `/weather Madrid` - Comando weather con argumento

## Crear un Comando Simple

Un comando es simplemente una función que sigue la firma `Command`:

```go
type Command func(context.Context, *Bot, *Message)
```

**Ejemplo básico:**

```go
func commandStart(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    welcome := "¡Bienvenido al bot!"
    if err := b.SendMessage(ctx, msg.Chat.ID, welcome); err != nil {
        log.Printf("Error: %v", err)
    }
}
```

## Registrar Comandos

Usa `CommandRegistry` para registrar tus comandos:

```go
// Crear el registro
commands := bot.NewCommandRegistry()

// Registrar comandos
commands.Register("start", commandStart)
commands.Register("help", commandHelp)

// Crear el bot con el registro de comandos
bot := bot.NewBot(token, bot.WithCommandRegistry(commands))
```

## Comandos con Argumentos

Puedes extraer argumentos del texto del mensaje:

```go
func commandWeather(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    // Dividir el texto en partes
    parts := strings.Fields(msg.Text)
    
    if len(parts) < 2 {
        b.SendMessage(ctx, msg.Chat.ID, "Uso: /weather <ciudad>")
        return
    }
    
    city := parts[1] // El primer argumento después del comando
    
    // Aquí harías la llamada a una API del clima
    response := fmt.Sprintf("El clima en %s es soleado", city)
    b.SendMessage(ctx, msg.Chat.ID, response)
}
```

## Comandos con Múltiples Argumentos

```go
func commandAdd(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    parts := strings.Fields(msg.Text)
    
    if len(parts) < 3 {
        b.SendMessage(ctx, msg.Chat.ID, "Uso: /add <nombre> <cantidad>")
        return
    }
    
    name := parts[1]
    quantity := parts[2]
    
    // Procesar los argumentos
    response := fmt.Sprintf("Agregado: %s x %s", name, quantity)
    b.SendMessage(ctx, msg.Chat.ID, response)
}
```

## Acceder a Información del Usuario

Puedes acceder a información del usuario que envió el comando:

```go
func commandInfo(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    user := msg.From
    
    info := fmt.Sprintf(
        "ID: %d\nNombre: %s %s\nUsername: @%s",
        user.ID,
        user.FirstName,
        user.LastName,
        user.Username,
    )
    
    b.SendMessage(ctx, msg.Chat.ID, info)
}
```

## Diferentes Comportamientos por Tipo de Chat

Puedes hacer que los comandos se comporten diferente según el tipo de chat:

```go
func commandStats(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    chat := msg.Chat
    
    switch chat.Type {
    case "private":
        b.SendMessage(ctx, chat.ID, "Estadísticas personales...")
    case "group", "supergroup":
        b.SendMessage(ctx, chat.ID, "Estadísticas del grupo...")
    case "channel":
        b.SendMessage(ctx, chat.ID, "Los canales no soportan este comando")
    }
}
```

## Comandos con Validación

Valida los argumentos antes de procesarlos:

```go
func commandDivide(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    parts := strings.Fields(msg.Text)
    
    if len(parts) != 3 {
        b.SendMessage(ctx, msg.Chat.ID, "Uso: /divide <a> <b>")
        return
    }
    
    a, err1 := strconv.ParseFloat(parts[1], 64)
    b, err2 := strconv.ParseFloat(parts[2], 64)
    
    if err1 != nil || err2 != nil {
        b.SendMessage(ctx, msg.Chat.ID, "Error: números inválidos")
        return
    }
    
    if b == 0 {
        b.SendMessage(ctx, msg.Chat.ID, "Error: división por cero")
        return
    }
    
    result := a / b
    response := fmt.Sprintf("Resultado: %.2f", result)
    b.SendMessage(ctx, msg.Chat.ID, response)
}
```

## Comandos que Usan Context para Timeouts

Usa el contexto para operaciones que pueden tardar:

```go
func commandSlowOperation(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    // Crear un contexto con timeout para esta operación
    opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    // Realizar una operación que puede tardar
    result := doSlowOperation(opCtx)
    
    if opCtx.Err() == context.DeadlineExceeded {
        b.SendMessage(ctx, msg.Chat.ID, "La operación tardó demasiado")
        return
    }
    
    b.SendMessage(ctx, msg.Chat.ID, result)
}
```

## Comandos con Estado

Puedes mantener estado usando variables de paquete o estructuras:

```go
type BotState struct {
    userSessions map[int64]*Session
    mu           sync.RWMutex
}

type Session struct {
    Step    string
    Data    map[string]interface{}
}

var state = &BotState{
    userSessions: make(map[int64]*Session),
}

func commandStart(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    state.mu.Lock()
    state.userSessions[msg.From.ID] = &Session{
        Step: "started",
        Data: make(map[string]interface{}),
    }
    state.mu.Unlock()
    
    b.SendMessage(ctx, msg.Chat.ID, "Sesión iniciada")
}
```

## Manejo de Comandos No Encontrados

El bot automáticamente ignora comandos no registrados. Si quieres responder a comandos desconocidos, puedes hacerlo en el handler de mensajes:

```go
func handleMessage(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    if strings.HasPrefix(msg.Text, "/") {
        // Es un comando, pero no fue encontrado
        b.SendMessage(ctx, msg.Chat.ID, "Comando no reconocido. Usa /help para ver los comandos disponibles.")
    }
}
```

**Nota:** Actualmente el bot procesa comandos automáticamente. Para manejar comandos no encontrados, necesitarías extender la funcionalidad del bot.

## Ejemplo Completo

Aquí tienes un ejemplo completo con múltiples comandos:

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
    welcome := "¡Bienvenido! Usa /help para ver los comandos disponibles."
    b.SendMessage(ctx, msg.Chat.ID, welcome)
}

func commandHelp(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    help := `Comandos disponibles:
/start - Iniciar el bot
/help - Mostrar esta ayuda
/time - Mostrar la hora actual
/echo <texto> - Repetir el texto`
    
    b.SendMessage(ctx, msg.Chat.ID, help)
}

func commandTime(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    now := time.Now().Format("2006-01-02 15:04:05")
    response := fmt.Sprintf("Hora actual: %s", now)
    b.SendMessage(ctx, msg.Chat.ID, response)
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
    commands := bot.NewCommandRegistry()
    commands.Register("start", commandStart)
    commands.Register("help", commandHelp)
    commands.Register("time", commandTime)
    commands.Register("echo", commandEcho)
    
    b := bot.NewBot(token, bot.WithCommandRegistry(commands))

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

## Mejores Prácticas

1. **Maneja errores**: Siempre verifica los errores al enviar mensajes
2. **Valida argumentos**: Verifica que los argumentos sean válidos antes de usarlos
3. **Usa context**: Aprovecha el contexto para timeouts y cancelación
4. **Mensajes claros**: Proporciona mensajes de error y ayuda claros
5. **Thread-safe**: Si compartes estado entre comandos, usa mutex o channels
6. **Logging**: Registra errores y eventos importantes para debugging

