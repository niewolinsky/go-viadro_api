package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Document struct {
	Document_id int       `json:"document_id"`
	User_id     int       `json:"user_id"`
	Url_s3      string    `json:"url_s3"`
	Filetype    string    `json:"filetype"`
	Uploaded_at time.Time `json:"uploaded_at"`
	Title       string    `json:"title"`
	Tags        []string  `json:"tags"`
	Is_hidden   bool      `json:"is_hidden"`
}

type DocumentLayer struct {
	DB *pgxpool.Pool
}

func (d DocumentLayer) Delete(id int) error {
	query := `
		DELETE FROM documents
		WHERE document_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := d.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (d DocumentLayer) Insert(document *Document) error {
	query := `
		INSERT INTO documents (filetype, title, tags, is_hidden, url_s3, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING document_id, uploaded_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{document.Filetype, document.Title, document.Tags, document.Is_hidden, document.Url_s3, document.User_id}

	err := d.DB.QueryRow(ctx, query, args...).Scan(&document.Document_id, &document.Uploaded_at)
	if err != nil {
		return err
	}

	return nil
}

func (d DocumentLayer) Get(id int) (*Document, error) {
	query := `
		SELECT document_id, user_id, url_s3, filetype, uploaded_at, title, tags, is_hidden
		FROM documents
		WHERE document_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	document := Document{}

	err := d.DB.QueryRow(ctx, query, id).Scan(
		&document.Document_id,
		&document.User_id,
		&document.Url_s3,
		&document.Filetype,
		&document.Uploaded_at,
		&document.Title,
		&document.Tags,
		&document.Is_hidden,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &document, nil
}

func (d DocumentLayer) GetAll(title string, tags []string, owner *int, flag *int, filters Filters) ([]Document, FilterMetadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), document_id, user_id, url_s3, filetype, uploaded_at, title, tags, is_hidden
		FROM documents
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (tags @> $2 OR $2 = '{}')
		AND ($3::int IS NOT NULL OR is_hidden = false)
		AND ($3::int IS NULL OR user_id = $3)
		AND ($4::int IS NULL OR user_id != $4)
		ORDER BY %s %s, document_id ASC
		LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, tags, owner, flag, filters.limit(), filters.offset()}

	rows, err := d.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, FilterMetadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	documents := []Document{}

	for rows.Next() {
		document := Document{}
		err := rows.Scan(
			&totalRecords,
			&document.Document_id,
			&document.User_id,
			&document.Url_s3,
			&document.Filetype,
			&document.Uploaded_at,
			&document.Title,
			&document.Tags,
			&document.Is_hidden,
		)
		if err != nil {
			return nil, FilterMetadata{}, err
		}
		documents = append(documents, document)
	}
	if err = rows.Err(); err != nil {
		return nil, FilterMetadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return documents, metadata, nil
}

func (d DocumentLayer) GetAllAdmin(title string, tags []string, filters Filters) ([]Document, FilterMetadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), document_id, user_id, url_s3, filetype, uploaded_at, title, tags, is_hidden
		FROM documents
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (tags @> $2 OR $2 = '{}')
		ORDER BY %s %s, document_id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, tags, filters.limit(), filters.offset()}

	rows, err := d.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, FilterMetadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	documents := []Document{}

	for rows.Next() {
		document := Document{}
		err := rows.Scan(
			&totalRecords,
			&document.Document_id,
			&document.User_id,
			&document.Url_s3,
			&document.Filetype,
			&document.Uploaded_at,
			&document.Title,
			&document.Tags,
			&document.Is_hidden,
		)
		if err != nil {
			return nil, FilterMetadata{}, err
		}
		documents = append(documents, document)
	}
	if err = rows.Err(); err != nil {
		return nil, FilterMetadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return documents, metadata, nil
}

func (d DocumentLayer) ToggleVisibility(id int) (*Document, error) {
	query := `
		UPDATE documents
		SET is_hidden = NOT is_hidden
		WHERE document_id = $1
		RETURNING document_id, url_s3, filetype, uploaded_at, title, tags, is_hidden
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	document := Document{}

	err := d.DB.QueryRow(ctx, query, id).Scan(
		&document.Document_id,
		&document.Url_s3,
		&document.Filetype,
		&document.Uploaded_at,
		&document.Title,
		&document.Tags,
		&document.Is_hidden,
	)
	if err != nil {
		return nil, err
	}

	return &document, nil
}
