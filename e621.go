package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var E621_USENAME = os.Getenv("E621_USERNAME")
var E621_API_KEY = os.Getenv("E621_API_KEY")

// Post represents a single e621 post
type Post struct {
	// updatedAt string `json:"updated_at"`
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	Score     struct {
		Up    int `json:"up"`
		Down  int `json:"down"`
		Total int `json:"total"`
	} `json:"score"`
	Tags struct {
		General   []string `json:"general"`
		Species   []string `json:"species"`
		Character []string `json:"character"`
		Artist    []string `json:"artist"`
		Meta      []string `json:"meta"`
		Copyright []string `json:"copyright"`
	} `json:"tags"`
	File struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Ext    string `json:"ext"`
		Size   int    `json:"size"`
		MD5    string `json:"md5"`
		URL    string `json:"url"`
	} `json:"file"`
	Preview struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		URL    string `json:"url"`
	} `json:"preview"`
	Sample struct {
		Has          bool   `json:"has"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		URL          string `json:"url"`
		ALT          string `json:"alt"`
		Alternatives struct {
			Has      bool    `json:"has"`
			Original Samples `json:"original"`
			Variants struct {
				Mp4 Samples `json:"mp4"`
			} `json:"variants"`
			Samples struct {
				S480p  Samples `json:"480p"`
				S720p  Samples `json:"720p"`
				S1080p Samples `json:"1080p"`
			} `json:"samples"`
		} `json:"alternatives"`
	} `json:"sample"`
	Rating        string   `json:"rating"`
	Description   string   `json:"description"`
	Sources       []string `json:"sources"`
	Relationships struct {
		ParentID          int   `json:"parent_id"`
		HasChildren       bool  `json:"has_children"`
		HasActiveChildren bool  `json:"has_active_children"`
		Children          []int `json:"children"`
	} `json:"relationships"`
}

type Samples struct {
	Fps    int    `json:"fps"`
	Codec  string `json:"codec"`
	Size   int    `json:"size"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Url    string `json:"url"`
}

type PostsResponse struct {
	Posts []Post `json:"posts"`
}

type E621Client struct {
	BaseURL   string
	UserAgent string
	Client    *http.Client
}

func NewE621Client(userAgent string) *E621Client {
	return &E621Client{
		BaseURL:   "https://e621.net",
		UserAgent: userAgent,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *E621Client) SearchPosts(tags []string, limit int, page int) (*PostsResponse, error) {
	endpoint := fmt.Sprintf("%s/posts.json", c.BaseURL)

	params := url.Values{}

	if len(tags) > 0 {
		tagsQuery := strings.Builder{}
		for i, tag := range tags {
			if i > 0 {
				tagsQuery.WriteByte(' ')
			}
			tagsQuery.WriteString(tag)
		}
		params.Add("tags", tagsQuery.String())
	}

	if limit > 0 {
		if limit > 320 {
			limit = 320
		}
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	if page > 0 {
		params.Add("page", fmt.Sprintf("%d", page))
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", E621_USENAME, E621_API_KEY))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var postsResp PostsResponse
	if err := json.NewDecoder(resp.Body).Decode(&postsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &postsResp, nil
}
