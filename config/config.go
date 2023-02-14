package config

import (
	"context"
	"flag"
	"fmt"
	"viadro_api/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Version string
	Port    int
	Env     string
	Db      struct {
		Dsn string
	}
}

func openPostgreDb(cfg Config) (*pgxpool.Pool, error) {
	fmt.Println(cfg.Db.Dsn)
	dbpool, err := pgxpool.New(context.Background(), cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

func InitConfig() (*pgxpool.Pool, Config) {
	cfg := Config{}

	flag.IntVar(&cfg.Port, "port", 4000, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|production)")
	flag.StringVar(&cfg.Db.Dsn, "db-dsn", "postgres://viadro:haslo456@localhost/viadro_db?sslmode=disable", "PostgreSQL DSN")
	flag.Parse()

	db_postgre, err := openPostgreDb(cfg)
	if err != nil {
		logger.LogFatal("failed opening database", err)
	}
	logger.LogInfo("database connection established")

	return db_postgre, cfg
}
