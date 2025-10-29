package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

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

func mdParse(s string) string {
	return bot.EscapeMarkdownUnescaped(bot.EscapeMarkdown(s))
}

func featured_handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("LOG: @%s entered /featured", update.Message.From.Username)
	dbClient := NewDerpibooruClient()
	response, dbErr := dbClient.getFeaturedImage()

	if dbErr != nil {
		log.Printf("ERROR: error while retreiving featured image: %v\n", dbErr)
	}
	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:     update.Message.Chat.ID,
		Photo:      &models.InputFileString{Data: response.Image.Representations.Full},
		HasSpoiler: response.Image.Spoilered,
		Caption:    bot.EscapeMarkdown(response.Image.Description) + fmt.Sprintf("\n\n[View on Derpibooru](https://derpibooru.org/images/%d)", response.Image.ID) + fmt.Sprintf("\nTags: %s", strings.Join(response.Image.Tags[:5], ", ")),
		ParseMode:  models.ParseModeMarkdown,
	})
	// _, err := b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByURL(response.Image.Representations.Full), &gotgbot.SendPhotoOpts{
	// 	Caption: caption,
	// 	// ParseMode:  "MarkdownV2",
	// 	HasSpoiler: response.Image.Spoilered,
	// })
	// if err != nil {
	// 	return fmt.Errorf("ERROR: could not send image: %v\n", err)
	// }
}

func inline_handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.InlineQuery == nil {
		return
	}
	dbClient := NewDerpibooruClient()
	page := 1
	query := update.InlineQuery.Query

	if strings.Contains(query, "#") {
		parts := strings.Split(query, "#")
		if len(parts) == 2 {
			if pageNum, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				page = pageNum
				query = strings.TrimSpace(parts[0])
			}
		}
	}

	log.Printf("QUERY IS %s\n\n\n", query)
	results, dbErr := dbClient.SearchImages(query, page, 25)
	if dbErr != nil {
		log.Printf("ERROR: error while searching images: %v\n", dbErr)
		return
	}
	var imageResults []*models.InlineQueryResultPhoto

	if len(results.Images) == 0 {
		log.Printf("LOG: no images found for query: %s\n", update.InlineQuery.Query)
	} else {
		for _, image := range results.Images {
			maxLen := 100
			if len(image.Description) > maxLen {
				image.Description = fmt.Sprintf("%s...", image.Description[:maxLen])
			}
			log.Printf("\nhttps://derpibooru.org/images/%d\nDESCRIPTION: %s\n\n", image.ID, image.Description)
			imageResults = append(imageResults, &models.InlineQueryResultPhoto{
				ID:           strconv.Itoa(image.ID),
				PhotoURL:     image.Representations.Full,
				ThumbnailURL: image.Representations.Thumb,
				Caption:      mdParse(image.Description) + fmt.Sprintf("\n\n[View on Derpibooru](https://derpibooru.org/images/%d)", image.ID),
				ParseMode:    models.ParseModeMarkdown,
				Title:        query,
			})
			// DEBUG:
			// log.Printf("DEBUG IMAGE URL: %v", image)
		}
	}

	var inlineResults []models.InlineQueryResult
	for _, result := range imageResults {
		inlineResults = append(inlineResults, result)
	}

	_, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: update.InlineQuery.ID + bot.RandomString(10),
		Results:       inlineResults,
		IsPersonal:    *bot.True(),
	})
	if err != nil {
		log.Printf("\nANSWER INLINE QUERY ERROR: %v\n", err)
		return
	}

	// _, err := update.InlineQuery.Answer(b, inlineResults, &bot.AnswerInlineQueryOpts{
	// 	IsPersonal: true,
	// })

	// log.Printf("LOG (inline): %s", ctx.InlineQuery.Query)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: failed to send Inline Query: %w", err)
	// }
}
