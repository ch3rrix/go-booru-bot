package main

import (
	"log"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/inlinequery"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		log.Fatal("ERROR: could not get TG_BOT_TOKEN environment variable!")
	}

	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		log.Fatalf("ERROR: could not create bot: %v\n", err)
	}

	if !bot.SupportsInlineQueries {
		log.Fatal("ERROR: bot does not support inline queries - enable them in botfather first!")
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Printf("ERROR: error happened while handling update: %v", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	updater := ext.NewUpdater(dispatcher, nil)

	dispatcher.AddHandler(handlers.NewInlineQuery(inlinequery.All, inline))
	dispatcher.AddHandler(handlers.NewCommand("featured", featured))
	dispatcher.AddHandler(handlers.NewCommand("start", start))

	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout:        9,
			AllowedUpdates: []string{"inline_query", "message"},
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("PANIC: failed to start polling: " + err.Error())
	}
	log.Printf("LOG: %s has been started", bot.User.Username)

	updater.Idle()
}
