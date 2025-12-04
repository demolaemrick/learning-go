package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tage"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
	Version   int       `json:"version"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count"`
}
type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content, title, user_id, tags) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags)).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `SELECT id, content, title, user_id, tags, version, created_at, updated_at 
		FROM posts 
		WHERE id = $1
		LIMIT 1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch err {
		// case errors.Is(err, sql.ErrNoRows):
		// 	return nil, ErrNotFound
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
        UPDATE posts
        SET title = $1,
            content = $2,
            tags = $3,
			version = version + 1,
            updated_at = now()
        WHERE id = $4 AND version = $5
        RETURNING content, title, tags, version, updated_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		post.ID,
	).Scan(&post.Content, &post.Title, pq.Array(&post.Tags), &post.Version, &post.UpdatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, pagination Pagination) ([]PostWithMetadata, error) {
	query := `
		SELECT 
			p.id, p.content, p.title, p.user_id, p.tags, p.version, p.created_at, u.username,
			COUNT(c.id) AS comment_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON u.id = p.user_id
		JOIN followers f ON f.follower_id = p.user_id OR f.follower_id = $1
		WHERE 
			f.user_id = $1 AND
			(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
			(p.tags @> $5 OR $5 = '{}')		
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + pagination.Sort + `
		LIMIT $2 OFFSET $3
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID, pagination.Limit, pagination.Offset, pagination.Search, pq.Array(pagination.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feed := []PostWithMetadata{}

	for rows.Next() {
		var post PostWithMetadata
		err := rows.Scan(
			&post.ID,
			&post.Content,
			&post.Title,
			&post.UserID,
			pq.Array(&post.Tags),
			&post.Version,
			&post.CreatedAt,
			&post.User.Username,
			&post.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, post)
	}
	return feed, nil
}
