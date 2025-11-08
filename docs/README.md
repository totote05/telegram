# Documentación del Proyecto Telegram Bot

Esta documentación proporciona una guía completa para usar la librería `telegram` de Go para interactuar con la API de Telegram Bot.

## Índice

- [Guía de Inicio Rápido](./getting-started.md) - Comienza a usar la librería en minutos
- [Referencia de API](./api-reference.md) - Documentación completa de todas las funciones y tipos
- [Guía de Comandos](./commands.md) - Cómo crear y gestionar comandos del bot
- [Ejemplos](./examples.md) - Ejemplos de código prácticos
- [Arquitectura](./architecture.md) - Diseño interno y decisiones arquitectónicas
- [Testing](./testing.md) - Guía para escribir y ejecutar tests

## Características Principales

- ✅ Interfaz simple y fácil de usar
- ✅ Soporte para comandos personalizados
- ✅ Manejo de contexto para cancelación y timeouts
- ✅ Long polling para recibir actualizaciones
- ✅ Código limpio y bien testeado
- ✅ Arquitectura modular y extensible

## Requisitos

- Go 1.24.5 o superior
- Un token de bot de Telegram (obtén uno de [@BotFather](https://t.me/BotFather))

## Instalación

```bash
go get github.com/totote05/telegram
```

## Uso Básico

```go
package main

import (
    "context"
    "github.com/totote05/telegram/bot"
)

func main() {
    bot := bot.NewBot("YOUR_BOT_TOKEN")
    ctx := context.Background()
    bot.Start(ctx)
}
```

Para más información, consulta la [Guía de Inicio Rápido](./getting-started.md).

