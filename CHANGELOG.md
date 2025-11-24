# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [0.2.0]

### Added
- Integración del logger estructurado de `github.com/totote05/go-toolkit/pkg/logger`
- Patrón de opciones funcionales para configurar el bot
- `WithLogger(log *slog.Logger) BotOption` - Permite reutilizar la instancia de logger del consumidor
- `WithCommandRegistry(registry *CommandRegistry) BotOption` - Configura comandos al crear el bot
- Logging estructurado con atributos contextuales en todas las operaciones del bot
- Stack traces automáticos en errores
- Documentación completa de las nuevas opciones funcionales

### Changed
- `NewBot(token string)` ahora acepta opciones funcionales: `NewBot(token string, opts ...BotOption)`
- Mejorada la observabilidad con logging estructurado en todas las operaciones
- Actualizada toda la documentación para reflejar el nuevo patrón de opciones

### Removed
- `SetCommandRegistry(commandRegistry *CommandRegistry)` - Reemplazado por `NewBot(token, WithCommandRegistry(registry))`

## [0.1.0]

### Added
- Implementación inicial de la librería
- `NewBot(token string)` - Constructor principal del bot
- `Start(ctx context.Context) error` - Inicia el bot con long polling
- `SendMessage(ctx context.Context, chatID int64, text string) error` - Envía mensajes
- `GetMe(ctx context.Context) error` - Verifica el token del bot
- `CommandRegistry` - Sistema de registro y ejecución de comandos
- `NewCommandRegistry()` - Crea un nuevo registro de comandos
- `Register(command string, action Command)` - Registra un comando
- `Execute(ctx context.Context, bot *Bot, msg *Message) bool` - Ejecuta un comando
- Soporte completo para context.Context en todas las operaciones
- Manejo de shutdown graceful mediante cancelación de contexto
- Procesamiento concurrente de mensajes en goroutines
- Long polling con timeout configurable
- Tests completos con alta cobertura
- Documentación completa en `/docs`
