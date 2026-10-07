package main

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

func nhknewsHTMLBuilder(feed *gofeed.Feed) string {
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

		articleHTML.WriteString(
			fmt.Sprintf(
				"<h2>%s</h2>\n",
				item.Title,
			),
		)

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
	}

	articleHTML.WriteString(`
</body>
</html>
`)

	return articleHTML.String()
}
