package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"viadro_api/internal/data"
)

const version = "0.1.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config      config
	data_access data.Layers
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	cfg := config{}

	{
		flag.IntVar(&cfg.port, "port", 4000, "API server port")
		flag.StringVar(&cfg.env, "env", "development", "Environment (development|production)")
		flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

		flag.Parse()
	}

	db, err := openDB(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, "database connection pool established")
	defer db.Close()

	app := &application{
		config:      cfg,
		data_access: data.NewLayers(db),
	}

	err = app.serve()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
