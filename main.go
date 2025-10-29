package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(inline_handler),
		bot.WithErrorsHandler(func(err error) {
			log.Printf("BOT HANDLER ERROR: %v", err)
		}),
	}

	b, err := bot.New(os.Getenv("TG_BOT_TOKEN"), opts...)
	if err != nil {
		log.Fatalf("BOT INIT ERROR: %v", err)
	}

	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		if update.Message == nil {
			return false
		}
		for _, e := range update.Message.Entities {
			if e.Type == models.MessageEntityTypeBotCommand {
				cmd := update.Message.Text[e.Offset+1 : e.Offset+e.Length]

				if idx := strings.Index(cmd, "@"); idx != -1 {
					cmd = cmd[:idx]
				}
				if cmd == "featured" {
					return true
				}
			}
		}
		return false
	}, featured_handler)

	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, start_handler)
	b.RegisterHandlerMatchFunc(matchFunc, echo_handler)

	b.Start(ctx)
}
