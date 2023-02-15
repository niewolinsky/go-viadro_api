package data

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Layers struct {
	Documents DocumentLayer
	Users     UserLayer
}

func NewLayers(db *pgxpool.Pool) Layers {
	return Layers{
		Documents: DocumentLayer{DB: db},
		Users:     UserLayer{DB: db},
	}
}
