package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func trimDescription(description string) string {
	if len(description) > 80 {
		description = fmt.Sprintf("Description: %s", description[:80])
	} else if len(description) <= 80 && len(description) > 0 {
		description = fmt.Sprintf("Description: %s", description)
	}
	// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
	strings.ReplaceAll(description, `_`, `\_`)
	strings.ReplaceAll(description, `*`, `\*`)
	strings.ReplaceAll(description, `[`, `\[`)
	strings.ReplaceAll(description, `]`, `\]`)
	strings.ReplaceAll(description, `(`, `\(`)
	strings.ReplaceAll(description, `)`, `\)`)
	strings.ReplaceAll(description, `~`, `\~`)
	strings.ReplaceAll(description, `>`, `\>`)
	strings.ReplaceAll(description, `#`, `\#`)
	strings.ReplaceAll(description, `-`, `\-`)
	strings.ReplaceAll(description, `=`, `\=`)
	strings.ReplaceAll(description, `|`, `\|`)
	strings.ReplaceAll(description, `{`, `\{`)
	strings.ReplaceAll(description, `}`, `\}`)
	strings.ReplaceAll(description, `.`, `\.`)
	strings.ReplaceAll(description, `!`, `\!`)

	fmt.Printf("DEBUG: trimDescription: %s\n", description)

	if !strings.HasSuffix(description, "\n") {
		return description + "\n"
	} else {
		return description
	}
}

func displayImageInILQ(img Image, id string) gotgbot.InlineQueryResultPhoto {
	// description := trimDescription(img.Description)
	description := fmt.Sprintf(`[Page URL](https://derpibooru.org/images/%s)`, strconv.Itoa(img.ID))
	return gotgbot.InlineQueryResultPhoto{
		Id:           id,
		PhotoUrl:     img.Representations.Full,
		ThumbnailUrl: img.Representations.Thumb,
		Caption:      description,
		ParseMode:    "MarkdownV2",
	}
}

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Printf("LOG: @%s entered /start", ctx.EffectiveUser.Username)
	/*	` This is just for reference
	 *	*bold \*text*
	 *	_italic \*text_
	 *	__underline__
	 *	~strikethrough~
	 *	||spoiler||
	 *	*bold _italic bold ~italic bold strikethrough ||italic bold strikethrough spoiler||~ __underline italic bold___ bold*
	 *	[inline URL](http://www.example.com/)
	 *	[inline mention of a user](tg://user?id=123456789)
	 *	![👍](tg://emoji?id=5368324170671202286)
	 *	`
	 *	*/
	startText := fmt.Sprint(`Hi\! This is a simple bot for searching and sending images from [derpibooru\.org](https://derpibooru.org)` +
		"\nThis bot is made by @ch3rrix\n" +
		"Commands:" +
		"\n/start" + ` \- prints this message` +
		"\n/featured" + ` \- sends today's featured image`)
	_, err := b.SendMessage(ctx.EffectiveChat.Id, startText, &gotgbot.SendMessageOpts{
		ParseMode:       "MarkdownV2",
		ReplyParameters: &gotgbot.ReplyParameters{MessageId: ctx.EffectiveMessage.MessageId},
	})
	if err != nil {
		return fmt.Errorf("ERROR: could not handle /start function: %v\n", err)
	}
	return nil
}
func featured(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Printf("LOG: @%s entered /featured", ctx.EffectiveUser.Username)
	dbClient := NewDerpibooruClient()
	response, dbErr := dbClient.getFeaturedImage()
	// description := trimDescription(*&response.Image.Description)

	caption := fmt.Sprintf(`[Page URL](https://derpibooru.org/images/%s)`,
		//description,
		strconv.Itoa(response.Image.ID),
	)

	if dbErr != nil {
		log.Printf("ERROR: error while retreiving featured image: %v\n", dbErr)
	}
	_, err := b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByURL(response.Image.Representations.Full), &gotgbot.SendPhotoOpts{
		Caption:    caption,
		ParseMode:  "MarkdownV2",
		HasSpoiler: response.Image.Spoilered,
	})
	if err != nil {
		return fmt.Errorf("ERROR: could not send image: %v\n", err)
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
