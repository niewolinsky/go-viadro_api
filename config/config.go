package config

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"viadro_api/internal/logger"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/wneessen/go-mail"
)

type Config struct {
	Version string
	Port    int
	Env     string
	Db      struct {
		Dsn string
	}
	Cache struct {
		Dsn      string
		Password string
		Index    int
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
}

func initializeMailClient(cfg Config) (*mail.Client, error) {
	mail_client, err := mail.NewClient(cfg.Smtp.Host, mail.WithPort(cfg.Smtp.Port), mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(cfg.Smtp.Username), mail.WithPassword(cfg.Smtp.Password))
	if err != nil {
		return nil, err
	}

	return mail_client, nil
}

func initializeS3Client() (*s3.Client, error) {
	aws_cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(aws_cfg)

	return client, nil
}

func openPostgreDb(cfg Config) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

func openRedis(cfg Config) (*redis.Client, error) {
	cache_client := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Dsn,
		Password: cfg.Cache.Password, // no password set
		DB:       cfg.Cache.Index,    // use default DB
	})
	_, err := cache_client.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	return cache_client, nil
}

func InitConfig() (*mail.Client, *s3.Client, *pgxpool.Pool, *redis.Client, Config) {
	cfg := Config{}

	err := godotenv.Load()
	if err != nil {
		logger.LogFatal("failed loading environment variables", err)
	}
	logger.LogInfo("environment variables loaded")

	//?APP
	APP_PORT, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	flag.IntVar(&cfg.Port, "port", APP_PORT, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|production)")

	//?POSTGRES
	flag.StringVar(&cfg.Db.Dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")

	//?REDIS
	flag.StringVar(&cfg.Cache.Dsn, "cache_dsn", os.Getenv("CACHE_DSN"), "Redis URI")
	flag.StringVar(&cfg.Cache.Password, "cache_password", os.Getenv("CACHE_PASSWORD"), "Redis Password")
	CACHE_INDEX, err := strconv.Atoi(os.Getenv("CACHE_INDEX"))
	if err != nil {
		log.Fatal(err)
	}
	flag.IntVar(&cfg.Cache.Index, "cache_index", CACHE_INDEX, "Redis Cache Number")

	//?SMTP
	flag.StringVar(&cfg.Smtp.Host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	SMTP_PORT, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	flag.IntVar(&cfg.Smtp.Port, "smtp-port", SMTP_PORT, "SMTP port")
	flag.StringVar(&cfg.Smtp.Username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.Smtp.Password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.Smtp.Sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	flag.Parse()
	logger.LogInfo("command line variables loaded")

	db_postgre, err := openPostgreDb(cfg)
	if err != nil {
		logger.LogFatal("failed opening database", err)
	}
	logger.LogInfo("database connection established")

	cache_client, err := openRedis(cfg)
	if err != nil {
		logger.LogFatal("failed opening cache", err)
	}
	logger.LogInfo("cache connection established")

	aws_s3_client, err := initializeS3Client()
	if err != nil {
		logger.LogFatal("failed initializing aws s3 manager", err)
	}
	logger.LogInfo("aws s3 client initialized")

	mail_client, err := initializeMailClient(cfg)
	if err != nil {
		logger.LogFatal("failed initializing mail client", err)
	}
	logger.LogInfo("mail client initialized")

	return mail_client, aws_s3_client, db_postgre, cache_client, cfg
}
