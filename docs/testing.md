# Guía de Testing

Esta guía explica cómo escribir y ejecutar tests para la librería y para bots que la usen.

## Ejecutar Tests

### Tests de la Librería

Para ejecutar todos los tests de la librería:

```bash
go test ./bot/...
```

Para ejecutar con cobertura:

```bash
go test -cover ./bot/...
```

Para ver cobertura detallada:

```bash
go test -coverprofile=coverage.out ./bot/...
go tool cover -html=coverage.out
```

### Tests con Verbosidad

```bash
go test -v ./bot/...
```

## Estructura de Tests

Los tests están organizados en archivos `*_test.go` junto al código fuente:

- `bot_test.go`: Tests para la funcionalidad principal del bot
- `command_test.go`: Tests para el sistema de comandos

## Estrategias de Testing

### 1. Mocking de HTTP

Para testear las interacciones con la API de Telegram, se usa `httptest.NewServer`:

```go
func TestBot_SendMessage(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verificar el método HTTP
        if r.Method != http.MethodPost {
            t.Errorf("expected POST, got %s", r.Method)
        }
        
        // Verificar headers
        if r.Header.Get("Content-Type") != "application/json" {
            t.Errorf("expected Content-Type application/json")
        }
        
        // Decodificar el request
        var req SendMessageRequest
        json.NewDecoder(r.Body).Decode(&req)
        
        // Verificar datos
        if req.ChatID != 123 {
            t.Errorf("expected ChatID 123, got %d", req.ChatID)
        }
        
        // Responder
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
    }))
    defer server.Close()
    
    // Crear bot con URL mockeada
    bot := &Bot{
        token:     "test-token",
        client:    &http.Client{Timeout: 5 * time.Second},
        apiBaseURL: server.URL + "/bot%s/%s",
    }
    
    // Ejecutar test
    ctx := context.Background()
    err := bot.SendMessage(ctx, 123, "Hello")
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### 2. Table-Driven Tests

Para testear múltiples casos, usa table-driven tests:

```go
func TestCommandRegistry_Execute(t *testing.T) {
    tests := []struct {
        name          string
        msgText       string
        command       string
        handlerExists bool
        wantExecuted  bool
    }{
        {
            name:          "execute existing command",
            msgText:       "/start",
            command:       "start",
            handlerExists: true,
            wantExecuted:  true,
        },
        {
            name:          "non-existent command",
            msgText:       "/unknown",
            command:       "start",
            handlerExists: true,
            wantExecuted:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            registry := NewCommandRegistry()
            
            callCount := 0
            if tt.handlerExists {
                handler := func(ctx context.Context, bot *Bot, msg *Message) {
                    callCount++
                }
                registry.Register(tt.command, handler)
            }
            
            msg := &Message{Text: tt.msgText, Chat: &Chat{ID: 123}}
            bot := NewBot("test-token")
            ctx := context.Background()
            
            executed := registry.Execute(ctx, bot, msg)
            
            if executed != tt.wantExecuted {
                t.Errorf("expected executed=%v, got %v", tt.wantExecuted, executed)
            }
        })
    }
}
```

### 3. Testing de Context

Para testear el manejo de cancelación y timeouts:

```go
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
    cancel() // Cancelar inmediatamente
    
    _, err := bot.makeRequest(ctx, "getMe", nil)
    if err == nil {
        t.Error("expected error due to context cancellation, got nil")
    }
}
```

## Testing de Tu Bot

### Ejemplo: Test de Comando

```go
package main

import (
    "context"
    "testing"
    
    "github.com/totote05/telegram/bot"
)

func TestCommandStart(t *testing.T) {
    // Crear un servidor mock
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
    }))
    defer server.Close()
    
    // Crear bot con URL mockeada
    b := &bot.Bot{
        token:     "test-token",
        client:    &http.Client{Timeout: 5 * time.Second},
        apiBaseURL: server.URL + "/bot%s/%s",
    }
    
    // Crear mensaje de prueba
    msg := &bot.Message{
        Text: "/start",
        From: &bot.User{ID: 123, FirstName: "Test"},
        Chat: &bot.Chat{ID: 456},
    }
    
    // Ejecutar comando
    ctx := context.Background()
    commandStart(ctx, b, msg)
    
    // Verificaciones adicionales si es necesario
}
```

### Ejemplo: Test de Validación

```go
func TestCommandDivide(t *testing.T) {
    tests := []struct {
        name    string
        text    string
        wantErr bool
    }{
        {"valid division", "/divide 10 2", false},
        {"division by zero", "/divide 10 0", true},
        {"invalid numbers", "/divide a b", true},
        {"missing args", "/divide", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock server
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                if !tt.wantErr {
                    w.WriteHeader(http.StatusOK)
                    w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
                }
            }))
            defer server.Close()
            
            b := &bot.Bot{
                token:     "test-token",
                client:    &http.Client{Timeout: 5 * time.Second},
                apiBaseURL: server.URL + "/bot%s/%s",
            }
            
            msg := &bot.Message{
                Text: tt.text,
                Chat: &bot.Chat{ID: 123},
            }
            
            ctx := context.Background()
            commandDivide(ctx, b, msg)
        })
    }
}
```

## Mejores Prácticas

### 1. Aislamiento

Cada test debe ser independiente y no depender de otros tests:

```go
// ✅ Bueno: Cada test crea su propio bot
func TestSomething(t *testing.T) {
    bot := NewBot("test-token")
    // ...
}

// ❌ Malo: Compartir estado entre tests
var sharedBot = NewBot("test-token")
```

### 2. Limpieza

Siempre limpia recursos (servers, goroutines, etc.):

```go
server := httptest.NewServer(...)
defer server.Close() // ✅ Siempre cerrar
```

### 3. Verificaciones Claras

Usa mensajes de error descriptivos:

```go
// ✅ Bueno
if err != nil {
    t.Errorf("expected no error, got %v", err)
}

// ❌ Malo
if err != nil {
    t.Error("error")
}
```

### 4. Nombres Descriptivos

Nombra los tests de forma descriptiva:

```go
// ✅ Bueno
func TestBot_SendMessage_WithValidChatID(t *testing.T) { }

// ❌ Malo
func Test1(t *testing.T) { }
```

### 5. Sub-tests

Usa `t.Run()` para organizar tests relacionados:

```go
func TestCommandRegistry_Execute(t *testing.T) {
    t.Run("existing command", func(t *testing.T) {
        // ...
    })
    
    t.Run("non-existent command", func(t *testing.T) {
        // ...
    })
}
```

## Testing de Concurrencia

Para testear comportamiento concurrente:

```go
func TestCommandRegistry_Execute_Concurrent(t *testing.T) {
    registry := NewCommandRegistry()
    
    callCount := 0
    var mu sync.Mutex
    
    handler := func(ctx context.Context, bot *Bot, msg *Message) {
        mu.Lock()
        callCount++
        mu.Unlock()
    }
    
    registry.Register("start", handler)
    
    bot := NewBot("test-token")
    ctx := context.Background()
    
    // Ejecutar concurrentemente
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            msg := &Message{Text: "/start", Chat: &Chat{ID: 123}}
            registry.Execute(ctx, bot, msg)
        }()
    }
    
    wg.Wait()
    
    if callCount != 10 {
        t.Errorf("expected callCount=10, got %d", callCount)
    }
}
```

## Cobertura de Código

### Ver Cobertura

```bash
go test -coverprofile=coverage.out ./bot/...
go tool cover -func=coverage.out
```

### Cobertura HTML

```bash
go test -coverprofile=coverage.out ./bot/...
go tool cover -html=coverage.out
```

Esto abrirá un navegador mostrando el código coloreado según la cobertura.

### Meta de Cobertura

Apunta a al menos 80% de cobertura para código crítico. La librería actual tiene buena cobertura de:
- ✅ Creación y configuración del bot
- ✅ Todas las operaciones de API
- ✅ Sistema de comandos
- ✅ Manejo de errores
- ✅ Cancelación de contexto

## CI/CD

### GitHub Actions

Ejemplo de workflow para ejecutar tests:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - run: go test -v -coverprofile=coverage.out ./bot/...
      - run: go tool cover -func=coverage.out
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./bot/...
```

### Ejecutar Test Específico

```bash
go test -v -run TestBot_SendMessage ./bot/...
```

### Timeout

Si un test se cuelga:

```bash
go test -timeout 30s ./bot/...
```

## Recursos Adicionales

- [Go Testing Package](https://pkg.go.dev/testing)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

