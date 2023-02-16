package data

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Layers struct {
	Documents DocumentLayer
	Users     UserLayer
	Tokens    TokenLayer
}

func NewLayers(db *pgxpool.Pool) Layers {
	return Layers{
		Documents: DocumentLayer{DB: db},
		Users:     UserLayer{DB: db},
		Tokens:    TokenLayer{DB: db},
	}
}
