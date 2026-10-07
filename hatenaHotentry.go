package main

import (
	"fmt"
	"offline-hatena/hatena"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
)

func hatenaHotentryHTMLBuilder(feed *gofeed.Feed) string {
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
