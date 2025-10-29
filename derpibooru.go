package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var DB_API_KEY = os.Getenv("DB_API_KEY")

type Image struct {
	UploaderID          int       `json:"uploader_id"`
	Spoilered           bool      `json:"spoilered"`
	DuplicateOf         *int      `json:"duplicate_of"`
	DeletionReason      *string   `json:"deletion_reason"`
	ViewURL             string    `json:"view_url"`
	Size                int       `json:"size"`
	TagIDs              []int     `json:"tag_ids"`
	CommentCount        int       `json:"comment_count"`
	Height              int       `json:"height"`
	Faves               int       `json:"faves"`
	MimeType            string    `json:"mime_type"`
	Processed           bool      `json:"processed"`
	Tags                []string  `json:"tags"`
	Description         string    `json:"description"`
	Score               int       `json:"score"`
	Duration            float64   `json:"duration"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Uploader            string    `json:"uploader"`
	ThumbnailsGenerated bool      `json:"thumbnails_generated"`
	Downvotes           int       `json:"downvotes"`
	SourceURLs          []string  `json:"source_urls"`
	FirstSeenAt         time.Time `json:"first_seen_at"`
	TagCount            int       `json:"tag_count"`
	HiddenFromUsers     bool      `json:"hidden_from_users"`
	SourceURL           string    `json:"source_url"`
	Format              string    `json:"format"`
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	WilsonScore         float64   `json:"wilson_score"`
	Width               int       `json:"width"`
	SHA512Hash          string    `json:"sha512_hash"`
	Upvotes             int       `json:"upvotes"`
	AspectRatio         float64   `json:"aspect_ratio"`
	OrigSHA512Hash      string    `json:"orig_sha512_hash"`
	Intensities         struct {
		Nw float64 `json:"nw"`
		Ne float64 `json:"ne"`
		Sw float64 `json:"sw"`
		Se float64 `json:"se"`
	} `json:"intensities"`
	Animated        bool `json:"animated"`
	Representations struct {
		Full       string `json:"full"`
		Small      string `json:"small"`
		ThumbTiny  string `json:"thumb_tiny"`
		ThumbSmall string `json:"thumb_small"`
		Thumb      string `json:"thumb"`
		Medium     string `json:"medium"`
		Large      string `json:"large"`
		Tall       string `json:"tall"`
	} `json:"representations"`
	OrigSize int `json:"orig_size"`
}

type SearchResponse struct {
	Images []Image `json:"images"`
	Total  int     `json:"total"`
}

type FeaturedResponse struct {
	Image Image `json:"image"`
}
type DerpibooruClient struct {
	BaseURL string
	Client  *http.Client
}

func NewDerpibooruClient() *DerpibooruClient {
	return &DerpibooruClient{
		BaseURL: "https://derpibooru.org/api/v1/json",
		Client:  &http.Client{},
	}
}

func (dbClient *DerpibooruClient) SearchImages(query string, page int, perPage int) (*SearchResponse, error) {
	u, err := url.Parse(dbClient.BaseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("ERROR (client.SearchImages): failed to parse URL: %v", err)
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("key", DB_API_KEY)
	params.Set("per_page", strconv.Itoa(perPage))
	u.RawQuery = params.Encode()

	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0")
	request.Header.Set("User-Agent", "Go-Booru-Bot/1.0")

	response, err := dbClient.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", response.StatusCode)
	}

	// Parse the JSON response
	var searchResponse SearchResponse
	if err := json.NewDecoder(response.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &searchResponse, nil
}

func (c *DerpibooruClient) getFeaturedImage() (*FeaturedResponse, error) {
	u, err := url.Parse(c.BaseURL + "/images/featured")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}
	params := url.Values{}
	params.Set("key", DB_API_KEY)
	u.RawQuery = params.Encode()

	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to create request: %v", err)
	}
	// request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0")
	request.Header.Set("User-Agent", "Go-Booru-Bot/1.0")

	response, err := c.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to make request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERROR: API returned status code: %d", response.StatusCode)
	}

	var featuredResponse FeaturedResponse
	if err := json.NewDecoder(response.Body).Decode(&featuredResponse); err != nil {
		return nil, fmt.Errorf("ERROR: failed to decode response: %v", err)
	}
	return &featuredResponse, nil
}
