package data

import (
	"database/sql"
)

type Layers struct {
	Documents DocumentLayer
}

func NewLayers(db *sql.DB) Layers {
	return Layers{
		Documents: DocumentLayer{DB: db},
	}
}
