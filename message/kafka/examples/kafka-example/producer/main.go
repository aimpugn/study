package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogType string

const (
    LogTypeDebug = LogType("debug")
    LogTypeErr = LogType("err")
    LogTypeWarn = LogType("warn")
    LogTypeInfo = LogType("info")
)

type Message struct {
	Type    LogType `json:"log_type"`
	Message string  `json:"message"`
}

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	app := fiber.New()

	app.Post("/logging", func(ctx fiber.Ctx) error {
		message := new(Message)

		if err := ctx.Bind().Body(message); err != nil {
			return fiber.ErrBadRequest
		}

		var event *zerolog.Event
		switch message.Type {
		case LogTypeDebug:
			event = log.Debug()
		case LogTypeErr:
			event = log.Error()
		case LogTypeWarn:
			event = log.Warn()
		case LogTypeInfo:
			event = log.Info()
		default:
			event = log.Debug()
		}
		event.Msg(message.Message)

		return ctx.SendString(message.Message + " is logged")
	})

	if err := app.Listen("0.0.0.0:9090"); err != nil {
		panic(err)
	}
}
