package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func displayImageInILQ(img Image, id string) gotgbot.InlineQueryResultPhoto {
	if img.Representations.Thumb == "" {
		log.Printf("Thumbnail: https:%s\n", img.Representations.Thumb)
	}

	if img.Description != "" {
		desc := strings.ReplaceAll(img.Description, "\n", " ")
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
	}

	var description string
	if len(img.Description) > 80 {
		description = fmt.Sprintf("Description: %s", img.Description[:80])
	} else if len(img.Description) <= 80 && len(img.Description) > 0 {
		description = fmt.Sprintf("Description: %s", img.Description)
	}
	return gotgbot.InlineQueryResultPhoto{
		Id:           id,
		PhotoUrl:     img.Representations.Full,
		ThumbnailUrl: img.Representations.Thumb,
		Caption:      description,
	}
}

func featured(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("USER ENTERED /featured")
	dbClient := NewDerpibooruClient()
	response, dbErr := dbClient.getFeaturedImage()
	caption := fmt.Sprintf("Description: %s\nTags: %s\nViewURL: %s", response.Image.Description, strings.Join(response.Image.Tags, ", "), response.Image.ViewURL)
	if dbErr != nil {
		log.Printf("ERROR: error while retreiving featured image: %v\n", dbErr)
	}
	_, err := b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByURL(response.Image.Representations.Full), &gotgbot.SendPhotoOpts{
		Caption: caption,
	})
	if err != nil {
		return fmt.Errorf("ERROR: could not send image: %v", err)
	}
	return nil
}
func inline(b *gotgbot.Bot, ctx *ext.Context) error {
	dbClient := NewDerpibooruClient()
	page := 1
	query := ctx.InlineQuery.Query
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
	}
	var inlineResults []gotgbot.InlineQueryResult
	if len(results.Images) == 0 {
		log.Printf("LOG: no images found for query: %s\n", ctx.InlineQuery.Query)
	} else {
		for i, image := range results.Images {
			inlineResults = append(inlineResults, displayImageInILQ(image, fmt.Sprintf("imline-query-string-id-%s", strconv.Itoa(i))))
		}
	}

	_, err := ctx.InlineQuery.Answer(b, inlineResults, &gotgbot.AnswerInlineQueryOpts{
		IsPersonal: true,
	})
	log.Printf("LOG (inline): %s", ctx.InlineQuery.Query)
	if err != nil {
		return fmt.Errorf("ERROR: failed to send Inline Query: %w", err)
	}

	return nil
}
