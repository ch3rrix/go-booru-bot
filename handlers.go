package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func echo_handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      update.Message.Text,
		ParseMode: models.ParseModeMarkdown,
	})
}

func matchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "hello"
}

func start_handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: bot.EscapeMarkdown(fmt.Sprintf(`Hello, %s\!`,
			update.Message.From.FirstName)),
		ParseMode: models.ParseModeMarkdown,
	})
}

func featured_handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("\nLOG: @%s entered /featured\n----------\n", update.Message.From.Username)
	dbClient := NewDerpibooruClient()
	response, dbErr := dbClient.getFeaturedImage()

	if dbErr != nil {
		log.Printf("\nERROR: could not get featured image: %v\n---------\n", dbErr)
	}
	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:     update.Message.Chat.ID,
		Photo:      &models.InputFileString{Data: response.Image.Representations.Full},
		HasSpoiler: response.Image.Spoilered,
		Caption: bot.EscapeMarkdown(response.Image.Description) +
			fmt.Sprintf("\n\n[View on Derpibooru](https://derpibooru.org/images/%d)", response.Image.ID) +
			fmt.Sprintf("\nTags: %s", strings.Join(response.Image.Tags[:5], ", ")),
		ParseMode: models.ParseModeMarkdown,
	})
}

func inline_handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	if update.InlineQuery == nil {
		return
	}

	inlineDebouncer := NewDebouncer()
	userID := update.InlineQuery.From.ID

	inlineDebouncer.Do(userID, 200*time.Millisecond, func() {
		process_inline_query(ctx, b, update, update.InlineQuery.Query)
	})

}
func process_inline_query(ctx context.Context, b *bot.Bot, update *models.Update, query string) {
	page := 1
	if update.InlineQuery.Offset != "" {
		page, _ = strconv.Atoi(update.InlineQuery.Offset)
	}

	log.Printf("\n@%s: %s\n-----------", update.InlineQuery.From.Username, query)

	var results []*models.InlineQueryResultPhoto
	var videoResults []*models.InlineQueryResultVideo
	if q, ok := strings.CutPrefix(query, "e621"); ok { // e621
		if q == "" {
			return
		}
		query = q
		e621Client := NewE621Client("Go-Booru-Bot/1.0 (by ch3rrix on e621.net)")
		search, err := e621Client.SearchPosts(strings.Fields(query), 20, page)
		if err != nil {
			log.Printf("\nE621 QUERY ERROR: %v\n----------\n", err)
		}
		if len(search.Posts) == 0 {
			log.Printf("\nE621 QUERY: no images found for query: %s\n----------\n", update.InlineQuery.Query)
		} else {
			for _, post := range search.Posts {
				switch post.File.Ext {
				case "png", "jpg", "jpeg", "webp":
					post.Description = trimDescription(post.Description, 100)
					results = append(results, &models.InlineQueryResultPhoto{
						ID:           strconv.Itoa(post.ID),
						PhotoURL:     post.File.URL,
						ThumbnailURL: post.Preview.URL,
						Caption: mdParse(post.Description) +
							fmt.Sprintf("\n\n[View on e621](https://e621.net/posts/%d)", post.ID),
						ParseMode: models.ParseModeMarkdown,
						Title:     query,
					})
				case "mp4":
					videoResults = append(videoResults, getVideoResultFromE621(post, query, "video/mp4"))
				case "webm":
					if post.Sample.Alternatives.Has {
						videoResults = append(videoResults, getVideoResultFromE621(post, query, "video/mp4"))
					}
				}
			}
		}
	} else if q, ok := strings.CutPrefix(query, "db"); ok { // Derpibooru
		if q == "" {
			return
		}
		query = q
		dbClient := NewDerpibooruClient()
		search, err := dbClient.SearchImages(query, page, 20)
		if err != nil {
			log.Printf("\nDB QUERY ERROR: %v\n----------\n", err)
			return
		}

		if len(search.Images) == 0 {
			log.Printf("DB QUERY: no images found for query: %s\n----------\n", update.InlineQuery.Query)
			return
		}
		for _, image := range search.Images {
			image.Description = trimDescription(image.Description, 100)
			switch image.MimeType {
			case "image/jpeg", "image/jpg", "image/png", "image/webp":
				results = append(results, &models.InlineQueryResultPhoto{
					ID:           strconv.Itoa(image.ID),
					PhotoURL:     image.Representations.Full,
					ThumbnailURL: image.Representations.Thumb,
					Caption:      mdParse(image.Description) + fmt.Sprintf("\n\n[View on Derpibooru](https://derpibooru.org/images/%d)", image.ID),
					ParseMode:    models.ParseModeMarkdown,
					Title:        query,
				})
			} // switch
		} // for
		/* Derpibooru */
	} else {
		return
	}

	var inline []models.InlineQueryResult
	for _, result := range results {
		inline = append(inline, result)
	}
	for _, result := range videoResults {
		inline = append(inline, result)
	}

	nextOffset := strconv.Itoa(page + 1)

	_, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: update.InlineQuery.ID, /*+ bot.RandomString(6)*/
		Results:       inline,
		IsPersonal:    *bot.True(),
		CacheTime:     10,
		NextOffset:    nextOffset,
	})
	if err != nil {
		log.Printf("\nANSWER INLINE QUERY ERROR: %v\n----------\n", err)
		return
	}
}
