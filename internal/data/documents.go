package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Document struct {
	Document_id int64     `json:"document_id"`
	User_id     int64     `json:"user_id"`
	Url_s3      string    `json:"url_s3"`
	Filetype    string    `json:"filetype"`
	Created_at  time.Time `json:"created_at"`
	Title       string    `json:"title"`
	Tags        []string  `json:"tags"`
	Is_hidden   bool      `json:"is_hidden"`
}

type DocumentLayer struct {
	DB *pgxpool.Pool
}

func (d DocumentLayer) Delete(id int64) error {
	query := `
	DELETE FROM documents
	WHERE document_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := d.DB.QueryRow(ctx, query, id)
	fmt.Println("Result:", result)

	return nil
}

func (d DocumentLayer) Insert(document *Document) error {
	query := `
		INSERT INTO documents (filetype, title, tags, is_hidden, url_s3)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING document_id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{document.Filetype, document.Title, document.Tags, document.Is_hidden, document.Url_s3}

	return d.DB.QueryRow(ctx, query, args...).Scan(&document.Document_id, &document.Created_at)
}

func (d DocumentLayer) Get(id int64) (*Document, error) {
	query := `
		SELECT document_id, user_id, url_s3, filetype, created_at, title, tags, is_hidden
		FROM documents
		WHERE document_id = $1
		`

	document := Document{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := d.DB.QueryRow(ctx, query, id).Scan(
		&document.Document_id,
		&document.User_id,
		&document.Url_s3,
		&document.Filetype,
		&document.Created_at,
		&document.Title,
		&document.Tags,
		&document.Is_hidden,
	)
	if err != nil {
		return nil, err
	}

	return &document, nil
}

func (d DocumentLayer) GetAll() ([]Document, error) {
	query := `
		SELECT document_id, user_id, url_s3, filetype, created_at, title, tags, is_hidden
		FROM documents
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := d.DB.Query(ctx, query)
	if err != nil {
		fmt.Println("error 1")
		return nil, err
	}
	defer rows.Close()

	documents := []Document{}
	for rows.Next() {
		document := Document{}
		err := rows.Scan(
			&document.Document_id,
			&document.User_id,
			&document.Url_s3,
			&document.Filetype,
			&document.Created_at,
			&document.Title,
			&document.Tags,
			&document.Is_hidden,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		documents = append(documents, document)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("error 3")
		return nil, err
	}

	return documents, nil
}

func (d DocumentLayer) ToggleVisibility(id int64) (*Document, error) {
	query := `
	UPDATE documents
	SET is_hidden = NOT is_hidden
	WHERE document_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	document := Document{}

	//!error???? bug??
	_ = d.DB.QueryRow(ctx, query, id).Scan(
		&document.Document_id,
		&document.User_id,
		&document.Url_s3,
		&document.Filetype,
		&document.Created_at,
		&document.Title,
		&document.Tags,
		&document.Is_hidden,
	)

	fmt.Println(document)

	return &document, nil
}
