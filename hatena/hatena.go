package hatena

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type Bookmark struct {
	User      string   `json:"user"`
	Tags      []string `json:"tags"`
	Timestamp string   `json:"timestamp"`
	Comment   string   `json:"comment"`
}

type HatenaEntry struct {
	EID       string     `json:"eid"`
	Bookmarks []Bookmark `json:"bookmarks"`
}

type HatenaStar struct {
	Name  string `json:"name"`
	Quote string `json:"quote"`
	Count int    `json:"count"`
}

type ColoredStars struct {
	Color string       `json:"color"`
	Stars []HatenaStar `json:"stars"`
}

type HatenaStarEntry struct {
	URI          string         `json:"uri"`
	Stars        []HatenaStar   `json:"stars"`
	ColoredStars []ColoredStars `json:"colored_stars"`
}

type HatenaStarResponse struct {
	Entries []HatenaStarEntry `json:"entries"`
}

type BookmarkWithCommentURL struct {
	Bookmark
	CommentURL string
}

type Star struct {
	User  string `json:"user"`
	Count int    `json:"count"`
	Color string `json:"color"`
	Quote string `json:"quote"`
}

type BookmarkCommentWithStars struct {
	User       string `json:"user"`
	Comment    string `json:"comment"`
	Timestamp  string `json:"timestamp"`
	CommentURL string `json:"commentUrl"`
	StarCount  int    `json:"starCount"`
	Stars      []Star `json:"stars"`
}

const starBatchSize = 50

func getJSON[T any](client *http.Client, requestURL string, result *T) error {
	resp, err := client.Get(requestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

func GetBookmarkCommentsWithStars(
	targetURL string,
) ([]BookmarkCommentWithStars, error) {

	client := http.DefaultClient

	// ----------------------------------------
	// 1. はてなブックマークのエントリー情報取得
	// ----------------------------------------

	entryURL := "https://b.hatena.ne.jp/entry/json/?" +
		url.Values{
			"url": []string{targetURL},
		}.Encode()

	var entry HatenaEntry

	if err := getJSON(client, entryURL, &entry); err != nil {
		return nil, err
	}

	if entry.EID == "" {
		return nil, fmt.Errorf("Hatena Bookmark entry not found")
	}

	// ----------------------------------------
	// 2. コメントURLを作成
	// ----------------------------------------

	comments := make([]BookmarkWithCommentURL, 0, len(entry.Bookmarks))

	for _, bookmark := range entry.Bookmarks {
		userID := bookmark.User

		// JavaScript:
		// bookmark.timestamp.slice(0, 10).replaceAll("/", "")
		date := bookmark.Timestamp

		if len(date) >= 10 {
			date = strings.ReplaceAll(date[:10], "/", "")
		}

		commentURL := fmt.Sprintf(
			"https://b.hatena.ne.jp/%s/%s#bookmark-%s",
			userID,
			date,
			entry.EID,
		)

		comments = append(comments, BookmarkWithCommentURL{
			Bookmark:   bookmark,
			CommentURL: commentURL,
		})
	}

	// ----------------------------------------
	// 3. コメントを50件ずつ処理
	// ----------------------------------------

	starMap := make(map[string]HatenaStarEntry)

	for i := 0; i < len(comments); i += starBatchSize {
		end := i + starBatchSize
		if end > len(comments) {
			end = len(comments)
		}

		batch := comments[i:end]

		params := url.Values{}

		for _, comment := range batch {
			params.Add("uri", comment.CommentURL)
		}

		starURL := "https://s.hatena.com/entry.json?" + params.Encode()

		var starResponse HatenaStarResponse

		if err := getJSON(client, starURL, &starResponse); err != nil {
			return nil, err
		}

		for _, starEntry := range starResponse.Entries {
			starMap[starEntry.URI] = starEntry
		}
	}

	// ----------------------------------------
	// 4. コメントとスターを結合
	// ----------------------------------------

	result := make([]BookmarkCommentWithStars, 0, len(comments))

	for _, comment := range comments {
		starEntry, exists := starMap[comment.CommentURL]

		stars := make([]Star, 0)
		starCount := 0

		if exists {
			// 黄色スター
			for _, star := range starEntry.Stars {
				count := star.Count
				if count == 0 {
					count = 1
				}

				stars = append(stars, Star{
					User:  star.Name,
					Count: count,
					Color: "yellow",
					Quote: star.Quote,
				})

				starCount += count
			}

			// カラースター
			for _, colored := range starEntry.ColoredStars {
				for _, star := range colored.Stars {
					count := star.Count
					if count == 0 {
						count = 1
					}

					stars = append(stars, Star{
						User:  star.Name,
						Count: count,
						Color: colored.Color,
						Quote: star.Quote,
					})

					starCount += count
				}
			}
		}

		result = append(result, BookmarkCommentWithStars{
			User:       comment.User,
			Comment:    comment.Comment,
			Timestamp:  comment.Timestamp,
			CommentURL: comment.CommentURL,
			StarCount:  starCount,
			Stars:      stars,
		})
	}

	return result, nil
}

// GetTop5BookmarkCommentsWithStars は、
// スター数が1以上のコメントをスター数降順に並べ、上位5件を返します。
func GetTop5BookmarkCommentsWithStars(
	targetURL string,
) ([]BookmarkCommentWithStars, error) {

	comments, err := GetBookmarkCommentsWithStars(targetURL)
	if err != nil {
		return nil, err
	}

	// starCount > 0 のコメントだけ残す
	topComments := make([]BookmarkCommentWithStars, 0)

	for _, comment := range comments {
		if comment.StarCount > 0 {
			topComments = append(topComments, comment)
		}
	}

	// starCount の降順でソート
	sort.Slice(topComments, func(i, j int) bool {
		return topComments[i].StarCount > topComments[j].StarCount
	})

	// 上位5件
	if len(topComments) > 5 {
		topComments = topComments[:5]
	}

	return topComments, nil
}
