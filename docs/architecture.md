# Arquitectura

Este documento describe la arquitectura interna de la librería y las decisiones de diseño.

## Visión General

La librería está diseñada siguiendo principios de Clean Architecture y separación de responsabilidades. El código está organizado en un solo paquete `bot` que contiene toda la funcionalidad relacionada con bots de Telegram.

## Estructura del Proyecto

```
telegram/
├── bot/
│   ├── bot.go          # Lógica principal del bot
│   ├── bot_test.go     # Tests del bot
│   ├── command.go      # Sistema de comandos
│   ├── command_test.go # Tests de comandos
│   └── types.go        # Tipos de datos de Telegram
├── examples/
│   └── simple_bot/     # Ejemplo de uso
└── docs/               # Documentación
```

## Componentes Principales

### 1. Bot (`bot.go`)

El componente central que gestiona la comunicación con la API de Telegram.

**Responsabilidades:**
- Gestionar la conexión HTTP con la API de Telegram
- Realizar solicitudes a la API (getUpdates, sendMessage, getMe)
- Procesar actualizaciones recibidas
- Manejar el ciclo de vida del bot (start, stop)
- Gestionar el offset de actualizaciones para long polling

**Características de diseño:**
- Usa `http.Client` con timeout configurado (70 segundos)
- Implementa long polling con timeout de 60 segundos
- Procesa mensajes en goroutines separadas para no bloquear
- Soporta cancelación mediante `context.Context`
- Permite inyección de URL base para testing (`apiBaseURL`)

### 2. CommandRegistry (`command.go`)

Sistema de registro y ejecución de comandos.

**Responsabilidades:**
- Registrar comandos con sus handlers
- Extraer comandos del texto de mensajes
- Ejecutar handlers cuando se reciben comandos
- Manejar comandos con mención de bot (`@botname`)

**Características de diseño:**
- Usa un mapa interno para almacenar comandos
- Extrae comandos removiendo el prefijo `/` y `@botname` si está presente
- Retorna `bool` para indicar si el comando fue encontrado y ejecutado
- Thread-safe para ejecución concurrente

### 3. Tipos (`types.go`)

Definiciones de estructuras que representan entidades de la API de Telegram.

**Tipos principales:**
- `Update`: Actualización recibida de Telegram
- `Message`: Mensaje de Telegram
- `User`: Usuario de Telegram
- `Chat`: Chat (privado, grupo, canal)
- `Response`: Respuesta genérica de la API
- `SendMessageRequest`: Request para enviar mensajes

**Características de diseño:**
- Usa tags JSON para serialización/deserialización
- Campos opcionales marcados con `omitempty`
- Tipos anidados para representar relaciones (Message contiene User y Chat)

## Flujo de Datos

### Inicio del Bot

```
main() 
  → NewBot(token)
  → SetCommandRegistry(registry) [opcional]
  → Start(ctx)
    → GetMe() [verificación]
    → Loop:
      → getUpdates(ctx)
      → Procesar cada update
        → handleMessage(ctx, msg)
          → Si es comando: CommandRegistry.Execute()
          → Si es mensaje normal: responder automáticamente
```

### Procesamiento de Mensajes

1. **Recepción**: El bot recibe actualizaciones mediante `getUpdates()`
2. **Extracción**: Se extrae el mensaje de cada update
3. **Procesamiento**: Se llama a `handleMessage()` en una goroutine
4. **Detección de Comandos**: Si el mensaje empieza con `/`, se intenta ejecutar el comando
5. **Respuesta**: El handler del comando o el handler por defecto envía una respuesta

## Decisiones de Diseño

### 1. Uso de Context

**Decisión**: Usar `context.Context` extensivamente para todas las operaciones.

**Razón**: 
- Permite cancelación y timeouts
- Facilita el shutdown graceful
- Sigue las mejores prácticas de Go
- Permite propagación de valores (trazas, timeouts, etc.)

### 2. Procesamiento Concurrente

**Decisión**: Procesar cada mensaje en una goroutine separada.

**Razón**:
- No bloquea la recepción de nuevos mensajes
- Permite manejar múltiples mensajes simultáneamente
- Mejora la capacidad de respuesta del bot

**Consideración**: Los handlers deben ser thread-safe si acceden a estado compartido.

### 3. Long Polling

**Decisión**: Usar long polling en lugar de webhooks.

**Razón**:
- Más simple de implementar (no requiere servidor HTTP)
- No requiere configuración de dominio/SSL
- Adecuado para bots pequeños y medianos
- El timeout de 60 segundos balancea latencia y recursos

### 4. Inyección de Dependencias Limitada

**Decisión**: Permitir inyección de `apiBaseURL` pero no de `http.Client`.

**Razón**:
- `apiBaseURL` es necesario para testing
- `http.Client` tiene una configuración específica (timeout de 70s)
- Simplifica la API pública
- Si se necesita más control, se puede extender en el futuro

### 5. Sistema de Comandos Simple

**Decisión**: Sistema de comandos basado en funciones, no en interfaces.

**Razón**:
- Más simple y directo
- Menos abstracción innecesaria
- Fácil de usar para casos comunes
- Puede extenderse con interfaces si es necesario

## Manejo de Errores

### Estrategia General

1. **Errores de API**: Se retornan como errores envueltos con contexto
2. **Errores de Red**: Se propagan con información del contexto
3. **Errores de Serialización**: Se retornan con mensajes descriptivos
4. **Errores de Contexto**: Se manejan explícitamente (context.Canceled, context.DeadlineExceeded)

### Logging

- Se usa `log` estándar de Go
- Se registran errores importantes
- Se registra información del bot al iniciar (GetMe)
- Los errores en handlers de comandos deben ser manejados por el usuario

## Testing

### Estrategia

1. **Unit Tests**: Cada componente se testea de forma aislada
2. **HTTP Mocking**: Se usa `httptest.NewServer` para simular la API de Telegram
3. **Table-Driven Tests**: Se usan para cubrir múltiples casos
4. **Context Testing**: Se testea el manejo de cancelación y timeouts

### Cobertura

- Tests para creación del bot
- Tests para todas las operaciones de API
- Tests para el sistema de comandos
- Tests para manejo de errores
- Tests para cancelación de contexto

## Extensiones Futuras

### Posibles Mejoras

1. **Webhooks**: Soporte para webhooks además de long polling
2. **Middleware**: Sistema de middleware para comandos
3. **Rate Limiting**: Protección contra rate limits de la API
4. **Retry Logic**: Reintentos automáticos en caso de errores transitorios
5. **Métricas**: Instrumentación con OpenTelemetry
6. **Más Métodos de API**: Soporte para más métodos de la API de Telegram
7. **Tipos Adicionales**: Soporte para más tipos de actualizaciones (callbacks, inline queries, etc.)

## Dependencias

### Estándar de Go

- `context`: Para cancelación y timeouts
- `encoding/json`: Para serialización
- `net/http`: Para comunicación HTTP
- `strings`: Para manipulación de strings
- `time`: Para timeouts y fechas
- `log`: Para logging

### Externa

- Ninguna (solo biblioteca estándar de Go)

## Consideraciones de Rendimiento

1. **Long Polling**: El timeout de 60 segundos balancea latencia y carga del servidor
2. **Goroutines**: Cada mensaje se procesa en una goroutine, permitiendo concurrencia
3. **HTTP Client Reutilizado**: Se reutiliza el mismo cliente HTTP para todas las solicitudes
4. **Offset Management**: Se gestiona correctamente el offset para evitar procesar mensajes duplicados

## Seguridad

1. **Token**: El token se almacena en memoria, no se expone en logs
2. **Validación**: Se valida el token al iniciar (GetMe)
3. **Input Sanitization**: Los handlers de comandos deben validar sus propios inputs
4. **Context Timeouts**: Se usan timeouts para prevenir operaciones que se cuelguen

## Conclusión

La arquitectura está diseñada para ser:
- **Simple**: Fácil de entender y usar
- **Extensible**: Puede crecer con nuevas funcionalidades
- **Testeable**: Componentes aislados y mockeables
- **Robusta**: Manejo de errores y shutdown graceful
- **Idiomática**: Sigue las convenciones y mejores prácticas de Go

