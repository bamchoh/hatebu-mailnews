package main

import (
	"database/sql"
	"offline-hatena/hatena"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func OpenDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}

	if err := database.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) createTables() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			image_url TEXT,
			bookmark_count INTEGER,
			description TEXT,
			fetched_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_id INTEGER NOT NULL,
			user TEXT NOT NULL,
			comment TEXT,
			timestamp TEXT,
			comment_url TEXT,
			star_count INTEGER NOT NULL,
			FOREIGN KEY (article_id)
				REFERENCES articles(id)
				ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_comments_article_id
			ON comments(article_id);

		CREATE INDEX IF NOT EXISTS idx_articles_fetched_at
			ON articles(fetched_at);
	`)

	return err
}

func (d *Database) SaveArticle(
	title string,
	link string,
	imageURL string,
	bookmarkCount int,
	description string,
) (int64, error) {

	_, err := d.db.Exec(`
		INSERT INTO articles (
			link,
			title,
			image_url,
			bookmark_count,
			description,
			fetched_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(link) DO UPDATE SET
			title = excluded.title,
			image_url = excluded.image_url,
			bookmark_count = excluded.bookmark_count,
			description = excluded.description,
			fetched_at = excluded.fetched_at
	`,
		link,
		title,
		imageURL,
		bookmarkCount,
		description,
		time.Now().UTC().Format(time.RFC3339),
	)

	if err != nil {
		return 0, err
	}

	// INSERTではなくUPDATEになった場合でもIDを取得する
	var id int64

	err = d.db.QueryRow(
		`SELECT id FROM articles WHERE link = ?`,
		link,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *Database) SaveComments(
	articleID int64,
	comments []hatena.BookmarkCommentWithStars,
) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// 今回取得したコメントだけに更新するため、一旦削除
	_, err = tx.Exec(
		`DELETE FROM comments WHERE article_id = ?`,
		articleID,
	)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO comments (
			article_id,
			user,
			comment,
			timestamp,
			comment_url,
			star_count
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, comment := range comments {
		_, err := stmt.Exec(
			articleID,
			comment.User,
			comment.Comment,
			comment.Timestamp,
			comment.CommentURL,
			comment.StarCount,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
