package config

import (
	"context"
	"flag"
	"fmt"
	"viadro_api/internal/logger"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	Version string
	Port    int
	Env     string
	Db      struct {
		Dsn string
	}
}

func initializeAwsManager() (*manager.Uploader, error) {
	aws_cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(aws_cfg)
	uploader := manager.NewUploader(client)

	return uploader, nil
}

func openPostgreDb(cfg Config) (*pgxpool.Pool, error) {
	fmt.Println(cfg.Db.Dsn)
	dbpool, err := pgxpool.New(context.Background(), cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

func InitConfig() (*manager.Uploader, *pgxpool.Pool, Config) {
	cfg := Config{}

	err := godotenv.Load()
	if err != nil {
		logger.LogFatal("failed loading environment variables", err)
	}
	logger.LogInfo("environment variables loaded")

	flag.IntVar(&cfg.Port, "port", 4000, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|production)")
	flag.StringVar(&cfg.Db.Dsn, "db-dsn", "postgres://viadro:haslo456@localhost/viadro_db?sslmode=disable", "PostgreSQL DSN")
	flag.Parse()
	logger.LogInfo("command line variables loaded")

	db_postgre, err := openPostgreDb(cfg)
	if err != nil {
		logger.LogFatal("failed opening database", err)
	}
	logger.LogInfo("database connection established")

	aws_s3_manager, err := initializeAwsManager()
	if err != nil {
		logger.LogFatal("failed initializing aws s3 manager", err)
	}
	logger.LogInfo("aws s3 manager initialized")

	return aws_s3_manager, db_postgre, cfg
}
