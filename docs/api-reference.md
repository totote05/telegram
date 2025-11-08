# Referencia de API

Documentación completa de todas las funciones, tipos y estructuras disponibles en la librería.

## Paquete `bot`

### Tipos Principales

#### `Bot`

Estructura principal que representa un bot de Telegram.

```go
type Bot struct {
    // Campos privados
}
```

**Métodos:**

##### `NewBot(token string) *Bot`

Crea una nueva instancia del bot con el token proporcionado.

**Parámetros:**
- `token` (string): Token del bot obtenido de BotFather

**Retorna:**
- `*Bot`: Nueva instancia del bot

**Ejemplo:**
```go
bot := bot.NewBot("123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
```

##### `Start(ctx context.Context) error`

Inicia el bot y comienza a recibir actualizaciones mediante long polling. Este método bloquea hasta que el contexto sea cancelado.

**Parámetros:**
- `ctx` (context.Context): Contexto para controlar el ciclo de vida del bot

**Retorna:**
- `error`: Error si ocurre algún problema, o `context.Canceled` cuando se cancela el contexto

**Comportamiento:**
- Verifica el token llamando a `GetMe()`
- Inicia un loop de long polling para recibir actualizaciones
- Procesa mensajes en goroutines separadas
- Maneja shutdown graceful cuando el contexto es cancelado

**Ejemplo:**
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := bot.Start(ctx); err != nil && err != context.Canceled {
    log.Fatalf("Error: %v", err)
}
```

##### `SendMessage(ctx context.Context, chatID int64, text string) error`

Envía un mensaje de texto a un chat específico.

**Parámetros:**
- `ctx` (context.Context): Contexto para la solicitud
- `chatID` (int64): ID del chat destino
- `text` (string): Texto del mensaje a enviar

**Retorna:**
- `error`: Error si la solicitud falla

**Ejemplo:**
```go
err := bot.SendMessage(ctx, 123456789, "¡Hola desde el bot!")
if err != nil {
    log.Printf("Error enviando mensaje: %v", err)
}
```

##### `GetMe(ctx context.Context) error`

Obtiene información sobre el bot y la imprime en los logs.

**Parámetros:**
- `ctx` (context.Context): Contexto para la solicitud

**Retorna:**
- `error`: Error si la solicitud falla

**Nota:** Este método se llama automáticamente en `Start()` para verificar el token.

##### `SetCommandRegistry(commandRegistry *CommandRegistry)`

Establece el registro de comandos para el bot.

**Parámetros:**
- `commandRegistry` (*CommandRegistry): Registro de comandos a usar

**Ejemplo:**
```go
commands := bot.NewCommandRegistry()
commands.Register("start", commandStart)
bot.SetCommandRegistry(commands)
```

### Tipos de Datos

#### `Update`

Representa una actualización recibida de la API de Telegram.

```go
type Update struct {
    UpdateID int      `json:"update_id"`
    Message  *Message `json:"message,omitempty"`
}
```

#### `Message`

Representa un mensaje de Telegram.

```go
type Message struct {
    MessageID int    `json:"message_id"`
    From      *User  `json:"from,omitempty"`
    Chat      *Chat  `json:"chat"`
    Date      int64  `json:"date"`
    Text      string `json:"text,omitempty"`
}
```

**Campos:**
- `MessageID`: ID único del mensaje
- `From`: Usuario que envió el mensaje (puede ser nil)
- `Chat`: Información del chat donde se envió el mensaje
- `Date`: Timestamp Unix del mensaje
- `Text`: Contenido de texto del mensaje (puede estar vacío)

#### `User`

Representa un usuario de Telegram.

```go
type User struct {
    ID        int64  `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name,omitempty"`
    Username  string `json:"username,omitempty"`
}
```

#### `Chat`

Representa un chat de Telegram (puede ser privado, grupo o canal).

```go
type Chat struct {
    ID       int64  `json:"id"`
    Type     string `json:"type"` // "private", "group", "supergroup", "channel"
    Title    string `json:"title,omitempty"`
    Username string `json:"username,omitempty"`
}
```

#### `SendMessageRequest`

Estructura para enviar mensajes.

```go
type SendMessageRequest struct {
    ChatID int64  `json:"chat_id"`
    Text   string `json:"text"`
}
```

#### `Response`

Estructura genérica para respuestas de la API de Telegram.

```go
type Response struct {
    Ok          bool            `json:"ok"`
    Result      json.RawMessage `json:"result,omitempty"`
    Description string          `json:"description,omitempty"`
}
```

## Paquete `bot` - Comandos

### `CommandRegistry`

Registro para gestionar comandos del bot.

#### `NewCommandRegistry() *CommandRegistry`

Crea un nuevo registro de comandos vacío.

**Retorna:**
- `*CommandRegistry`: Nueva instancia del registro

#### `Register(command string, action Command)`

Registra un comando con su handler.

**Parámetros:**
- `command` (string): Nombre del comando (sin el prefijo `/`)
- `action` (Command): Función handler que se ejecutará cuando se invoque el comando

**Ejemplo:**
```go
commands.Register("start", func(ctx context.Context, b *Bot, msg *Message) {
    b.SendMessage(ctx, msg.Chat.ID, "¡Hola!")
})
```

**Nota:** Si registras el mismo comando dos veces, el segundo handler sobrescribirá al primero.

#### `Execute(ctx context.Context, bot *Bot, msg *Message) bool`

Ejecuta el comando correspondiente al mensaje, si existe.

**Parámetros:**
- `ctx` (context.Context): Contexto para la ejecución
- `bot` (*Bot): Instancia del bot
- `msg` (*Message): Mensaje que contiene el comando

**Retorna:**
- `bool`: `true` si el comando fue encontrado y ejecutado, `false` en caso contrario

**Comportamiento:**
- Extrae el comando del texto del mensaje (removiendo `/` y `@botname` si está presente)
- Busca el handler registrado para ese comando
- Si existe, lo ejecuta con el contexto, bot y mensaje
- Retorna `true` si el comando fue ejecutado, `false` si no se encontró

**Ejemplo:**
```go
// El bot llama automáticamente a Execute cuando recibe un mensaje que empieza con "/"
// Pero también puedes llamarlo manualmente:
executed := commands.Execute(ctx, bot, message)
if !executed {
    // Comando no encontrado
}
```

### `Command`

Tipo de función para handlers de comandos.

```go
type Command func(context.Context, *Bot, *Message)
```

**Parámetros:**
- `context.Context`: Contexto para la operación
- `*Bot`: Instancia del bot para enviar mensajes o realizar acciones
- `*Message`: Mensaje que contiene el comando

**Ejemplo:**
```go
func myCommand(ctx context.Context, b *bot.Bot, msg *bot.Message) {
    // Acceder a información del mensaje
    userID := msg.From.ID
    chatID := msg.Chat.ID
    text := msg.Text
    
    // Enviar respuesta
    b.SendMessage(ctx, chatID, "Comando ejecutado!")
}
```

## Constantes

### `apiURL`

URL base de la API de Telegram.

```go
const apiURL = "https://api.telegram.org/bot%s/%s"
```

### `timeout`

Timeout para long polling en segundos.

```go
const timeout = 60 // segundos
```

## Manejo de Errores

Todos los métodos que interactúan con la API de Telegram pueden retornar errores. Los errores comunes incluyen:

- **Errores de red**: Problemas de conectividad o timeouts
- **Errores de API**: Respuestas con `ok: false` de la API de Telegram
- **Errores de contexto**: Cuando el contexto es cancelado
- **Errores de serialización**: Problemas al codificar/decodificar JSON

Siempre verifica los errores retornados:

```go
if err := bot.SendMessage(ctx, chatID, text); err != nil {
    log.Printf("Error enviando mensaje: %v", err)
    // Manejar el error apropiadamente
}
```

## Context y Cancelación

La librería usa `context.Context` extensivamente para:

- **Cancelación**: Permite cancelar operaciones en curso
- **Timeouts**: Puedes usar `context.WithTimeout()` para limitar el tiempo de operaciones
- **Propagación de valores**: Puedes pasar valores a través del contexto

**Ejemplo con timeout:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := bot.SendMessage(ctx, chatID, text)
if err != nil {
    // Puede ser un timeout o error de red
}
```

