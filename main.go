package main

import (
	"fmt"
	"html"
	"log"
	"offline-hatena/hatena"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

func htmlBuilder(feed *gofeed.Feed) string {
	var articleHTML strings.Builder

	articleHTML.WriteString(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
`)

	for _, item := range feed.Items {
		if item.Link == "" {
			continue
		}

		comments, err := hatena.GetTop5BookmarkCommentsWithStars(item.Link)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"コメント取得エラー: %s: %v\n",
				item.Link,
				err,
			)
			comments = nil
		}

		// カスタムフィールド
		bookmarkCount := item.Extensions["hatena"]["bookmarkcount"]
		imageURL := item.Extensions["hatena"]["imageurl"]

		articleHTML.WriteString(
			fmt.Sprintf(
				"<h2>%s - ☆%s</h2>\n",
				item.Title,
				bookmarkCount[0].Value,
			),
		)

		if len(imageURL) > 0 {
			articleHTML.WriteString(
				fmt.Sprintf(
					`<img src="%s" alt="%s" style="max-width:100%%;height:auto;"><br><br>`+"\n",
					imageURL[0].Value,
					item.Title,
				),
			)
		}

		articleHTML.WriteString(
			fmt.Sprintf(
				"<p>%s</p>\n",
				item.Description,
			),
		)

		articleHTML.WriteString(
			fmt.Sprintf(
				`<p><a href="%s">記事を読む</a></p>`+"\n",
				item.Link,
			),
		)

		if len(comments) > 0 {
			articleHTML.WriteString(`
<h3>注目コメント</h3>
<ul>
`)

			for _, comment := range comments {
				articleHTML.WriteString(
					fmt.Sprintf(
						"<li>%s <strong>%s</strong> ★%d</li>\n",
						comment.Comment,
						comment.User,
						comment.StarCount,
					),
				)
			}

			articleHTML.WriteString(`
</ul>
<br>
`)
		}
	}

	articleHTML.WriteString(`
</body>
</html>
`)

	return articleHTML.String()
}

func markdownBuilder(feed *gofeed.Feed) string {
	var articleMarkdown strings.Builder

	articleMarkdown.WriteString(
		"# はてなブックマーク ホットエントリー（IT）\n\n",
	)

	for _, item := range feed.Items {
		if item.Link == "" {
			continue
		}

		comments, err := hatena.GetTop5BookmarkCommentsWithStars(item.Link)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"コメント取得エラー: %s: %v\n",
				item.Link,
				err,
			)
			comments = nil
		}

		// カスタムフィールド
		bookmarkCount := item.Extensions["hatena"]["bookmarkcount"]
		imageURL := item.Extensions["hatena"]["imageurl"]
		title := html.EscapeString(item.Title)
		description := html.EscapeString(item.Description)

		articleMarkdown.WriteString(
			fmt.Sprintf(
				"## %s - ☆%s\n\n",
				title,
				bookmarkCount[0].Value,
			),
		)

		if len(imageURL) > 0 {
			articleMarkdown.WriteString(
				fmt.Sprintf(
					"![%s](%s)\n\n",
					title,
					imageURL[0].Value,
				),
			)
		}

		articleMarkdown.WriteString(
			fmt.Sprintf(
				"%s\n\n",
				description,
			),
		)

		articleMarkdown.WriteString(
			fmt.Sprintf(
				"- URL: %s\n\n",
				item.Link,
			),
		)

		if len(comments) > 0 {
			articleMarkdown.WriteString(
				"<details>\n  <summary>注目コメント</summary>\n\n",
			)

			for _, comment := range comments {
				articleMarkdown.WriteString(
					fmt.Sprintf(
						"  - %s (**%s** ★%d)\n",
						comment.Comment,
						comment.User,
						comment.StarCount,
					),
				)
			}

			articleMarkdown.WriteString(
				"\n</details>\n\n",
			)
		}
	}

	return articleMarkdown.String()
}

func sendFeedMail(rssURL string, subject string) error {
	parser := gofeed.NewParser()

	feed, err := parser.ParseURL(rssURL)
	if err != nil {
		return err
	}

	subject = fmt.Sprintf(
		"%s (%s)",
		subject,
		time.Now().Format("2006/01/02"),
	)

	return sendMail(
		os.Getenv("GMAIL_ADDRESS"),
		subject,
		htmlBuilder(feed),
	)
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println(".env not found, using environment variables")
	}

	rssList := []map[string]string{
		{
			"url":     "https://b.hatena.ne.jp/hotentry/it.rss",
			"subject": "はてブ【カテゴリ: IT】",
		},
		{
			"url":     "https://b.hatena.ne.jp/hotentry/all.rss",
			"subject": "はてブ【カテゴリ: 総合】",
		},
	}

	for _, rss := range rssList {
		if err := sendFeedMail(rss["url"], rss["subject"]); err != nil {
			log.Fatal(err)
		}
	}
}
