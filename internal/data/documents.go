package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Document struct {
	Document_id int64  `json:"id"`
	Title       string `json:"title"`
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
		INSERT INTO documents (title)
		VALUES ($1)
		RETURNING document_id
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{document.Title}

	return d.DB.QueryRow(ctx, query, args...).Scan(&document.Document_id)
}

func (d DocumentLayer) GetAll() ([]Document, error) {
	query := `
		SELECT document_id, title
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
			&document.Title,
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
