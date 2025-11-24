# telegram

Una librería de Go simple y fácil de usar para interactuar con la API de Telegram Bot.

## 📋 Descripción

`telegram` es una librería de Go que proporciona una interfaz limpia y sencilla para crear bots de Telegram. Está diseñada siguiendo principios de Clean Architecture y ofrece una API intuitiva para manejar mensajes, comandos y actualizaciones de Telegram.

## ✨ Características

- ✅ **Interfaz simple y fácil de usar** - API intuitiva y directa
- ✅ **Soporte para comandos personalizados** - Sistema flexible de registro de comandos
- ✅ **Manejo de contexto** - Soporte completo para cancelación y timeouts
- ✅ **Long polling** - Recepción eficiente de actualizaciones
- ✅ **Código limpio y bien testeado** - Alta cobertura de tests
- ✅ **Arquitectura modular** - Diseño extensible y mantenible
- ✅ **Sin dependencias externas** - Solo biblioteca estándar de Go

## 📦 Instalación

```bash
go get github.com/totote05/telegram
```

## 🚀 Uso Rápido

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
        log.Printf("Error enviando mensaje: %v", err)
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
        log.Println("Cerrando...")
        cancel()
    }()

    if err := b.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error ejecutando bot: %v", err)
    }
}
```

## 📚 Documentación

Para documentación completa, ejemplos y guías detalladas, consulta la [documentación en `/docs`](./docs/README.md):

- [Guía de Inicio Rápido](./docs/getting-started.md) - Comienza a usar la librería en minutos
- [Referencia de API](./docs/api-reference.md) - Documentación completa de todas las funciones y tipos
- [Guía de Comandos](./docs/commands.md) - Cómo crear y gestionar comandos del bot
- [Ejemplos](./docs/examples.md) - Ejemplos de código prácticos
- [Arquitectura](./docs/architecture.md) - Diseño interno y decisiones arquitectónicas
- [Testing](./docs/testing.md) - Guía para escribir y ejecutar tests

## 🔧 Requisitos

- **Go 1.24.5** o superior
- Un **token de bot de Telegram** (obtén uno de [@BotFather](https://t.me/BotFather))

## 📝 Estado del Proyecto

**Versión:** 0.1.0 (En desarrollo)

Este proyecto está actualmente en desarrollo activo. Las características principales están implementadas y funcionando, pero pueden haber cambios en la API antes de alcanzar la versión 1.0.0.

## 🏗️ Estructura del Proyecto

```
telegram/
├── bot/              # Código fuente principal
│   ├── bot.go        # Lógica principal del bot
│   ├── command.go    # Sistema de comandos
│   └── types.go      # Tipos de datos de Telegram
├── examples/         # Ejemplos de uso
│   └── simple_bot/  # Bot de ejemplo
└── docs/            # Documentación completa
```

## 📄 Licencia

Este proyecto está licenciado bajo la [MIT License](./LICENSE).

## 👤 Autor

Desarrollado por [totote05](https://github.com/totote05)

## 🤝 Obtener Ayuda

- 📖 Consulta la [documentación completa](./docs/README.md)
- 💬 Abre un [issue](https://github.com/totote05/telegram/issues) para reportar bugs o solicitar características
- 📧 Para preguntas o soporte, puedes abrir una discusión en el repositorio

---

**Nota:** Este proyecto está en desarrollo activo. Si encuentras algún problema o tienes sugerencias, ¡no dudes en abrir un issue!
