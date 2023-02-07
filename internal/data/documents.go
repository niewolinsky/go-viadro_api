package data

import (
	"database/sql"
	"time"
)

type Document struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
}

type DocumentLayer struct {
	DB *sql.DB
}
