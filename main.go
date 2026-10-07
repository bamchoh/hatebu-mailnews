package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/go-trending"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/errgroup"
)

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

func sendFeedMail(rssURL string, subject string, contentBuilder func(*gofeed.Feed) string) error {
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
		contentBuilder(feed),
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

type RSSConfig struct {
	URL         string
	Subject     string
	HTMLBuilder func(*gofeed.Feed) string
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

	rssList := []RSSConfig{
		{
			URL:         "https://b.hatena.ne.jp/hotentry/it.rss",
			Subject:     "はてブ【カテゴリ: IT】",
			HTMLBuilder: hatenaHotentryHTMLBuilder,
		},
		{
			URL:         "https://b.hatena.ne.jp/hotentry/all.rss",
			Subject:     "はてブ【カテゴリ: 総合】",
			HTMLBuilder: hatenaHotentryHTMLBuilder,
		},
		{
			URL:         "https://news.web.nhk/n-data/conf/na/rss/cat0.xml",
			Subject:     "NHK News【カテゴリ: 総合】",
			HTMLBuilder: nhknewsHTMLBuilder,
		},
		{
			URL:         "https://news.web.nhk/n-data/conf/na/rss/cat2.xml",
			Subject:     "NHK News【カテゴリ: 暮らし】",
			HTMLBuilder: nhknewsHTMLBuilder,
		},
	}

	var g errgroup.Group

	for _, rss := range rssList {
		g.Go(func() error {
			return sendFeedMail(rss.URL, rss.Subject, rss.HTMLBuilder)
		})
	}

	return g.Wait()
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(run)
		return
	} else {
		if err := run(); err != nil {
			log.Fatal(err)
		}
	}
}
