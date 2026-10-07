package main

import (
	"fmt"
	"offline-hatena/hatena"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/go-trending"
	"github.com/aws/aws-lambda-go/lambda"
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

func githubTrendingHTMLBuilder(projects []trending.Project) string {
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

	for _, project := range projects {
		articleHTML.WriteString(
			fmt.Sprintf(
				"<h2><a href=\"%s\">%s</a></h2>\n",
				project.URL,
				project.Name,
			),
		)
		articleHTML.WriteString(
			fmt.Sprintf(
				"<p>%s</p>\n",
				project.Description,
			),
		)
		articleHTML.WriteString(
			fmt.Sprintf(
				"<p>★ %d</p>\n",
				project.Stars,
			),
		)
		if len(project.Language) > 0 {
			articleHTML.WriteString(
				fmt.Sprintf(
					"<p>言語: %s</p>\n",
					project.Language,
				),
			)
		}
	}

	articleHTML.WriteString(`
</body>
</html>
`)

	return articleHTML.String()
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

func githubTrending(dateRange, language string) error {
	trend := trending.NewTrending()

	// Show projects of today
	projects, err := trend.GetProjects(dateRange, language)
	if err != nil {
		return err
	}

	githubTrendingHTML := githubTrendingHTMLBuilder(projects)

	var subject string
	if language != "" {
		subject = fmt.Sprintf(
			"GitHub Trending (%s) (%s) (%s)",
			dateRange,
			language,
			time.Now().Format("2006/01/02"),
		)
	} else {
		subject = fmt.Sprintf(
			"GitHub Trending (%s) (Any) (%s)",
			dateRange,
			time.Now().Format("2006/01/02"),
		)
	}

	return sendMail(
		os.Getenv("GMAIL_ADDRESS"),
		subject,
		githubTrendingHTML,
	)
}

func run() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println(".env not found, using environment variables")
	}

	if err := githubTrending(trending.TimeToday, "go"); err != nil {
		return err
	}

	if err := githubTrending(trending.TimeToday, ""); err != nil {
		return err
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
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(run)
}
