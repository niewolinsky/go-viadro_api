package data

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Document struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
}

type DocumentLayer struct {
	DB *pgxpool.Pool
}
