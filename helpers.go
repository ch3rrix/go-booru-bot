package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func getVideoResultFromE621(post Post, query string, mimetype string) *models.InlineQueryResultVideo {
	artist := strings.Join(post.Tags.Artist, ", ")
	if artist != "" {
		artist = fmt.Sprintf("Artist: %s, ", artist)
	}
	title := trimDescription(artist+strings.Join(post.Tags.Character, ", "), 80)
	if title == "" {
		title = query
	}
	url := post.File.URL

	if strings.HasSuffix(url, ".mp4") {
		v := []string{
			post.Sample.Alternatives.Original.Url,
			post.Sample.Alternatives.Variants.Mp4.Url,
			post.Sample.Alternatives.Samples.S1080p.Url,
			post.Sample.Alternatives.Samples.S720p.Url,
			post.Sample.Alternatives.Samples.S480p.Url,
		}

		for _, u := range v {
			if strings.HasSuffix(u, ".mp4") {
				url = u
			}
		}
	} else {
		log.Printf("\nERROR while checking extension\nPost Preview: %v\nPost Sample: %v\n---------\n", post.Preview, post.Sample)
	}

	return &models.InlineQueryResultVideo{
		ID:           strconv.Itoa(post.ID),
		VideoURL:     url,
		ThumbnailURL: post.Preview.URL,
		Caption: mdParse(post.Description) +
			fmt.Sprintf("\n\n[View on e621](https://e621.net/posts/%d)", post.ID),
		ParseMode: models.ParseModeMarkdown,
		MimeType:  mimetype,
		Title:     title,
	}
}

func mdParse(s string) string {
	return bot.EscapeMarkdownUnescaped(bot.EscapeMarkdown(s))
}

func trimDescription(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return fmt.Sprintf("%s...\n", s[:max])
}
