package config

import (
	"context"
	"flag"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/wneessen/go-mail"
)

type configuration struct {
	version string
	port    string
	env     string
	db      struct {
		dsn string
	}
	cache struct {
		dsn      string
		password string
		index    int
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

func initializePostgresClient(cfg configuration) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

func initializeRedisClient(cfg configuration) (*redis.Client, error) {
	cache_client := redis.NewClient(&redis.Options{
		Addr:     cfg.cache.dsn,
		Password: cfg.cache.password, // no password set
		DB:       cfg.cache.index,    // use default DB
	})
	_, err := cache_client.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	return cache_client, nil
}

func initializeS3Client() (*s3.Client, error) {
	aws_cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	s3_client := s3.NewFromConfig(aws_cfg)

	return s3_client, nil
}

func initializeMailClient(cfg configuration) (*mail.Client, error) {
	mail_client, err := mail.NewClient(cfg.smtp.host, mail.WithPort(cfg.smtp.port), mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(cfg.smtp.username), mail.WithPassword(cfg.smtp.password))
	if err != nil {
		return nil, err
	}

	return mail_client, nil
}

func InitConfig() (*mail.Client, *s3.Client, *pgxpool.Pool, *redis.Client, string) {
	config := configuration{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed loading environment variables", err)
	}
	log.Info("environment variables loaded")

	//?APP
	flag.StringVar(&config.port, "port", os.Getenv("APP_PORT"), "application erver port")
	flag.StringVar(&config.version, "version", os.Getenv("APP_VERSION"), "application version")
	flag.StringVar(&config.env, "env", os.Getenv("APP_ENVIRONMENT"), "application environment")

	//?POSTGRES
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("POSTGRES_DSN"), "PostgreSQL DSN")

	//?REDIS
	flag.StringVar(&config.cache.dsn, "redis_dsn", os.Getenv("REDIS_DSN"), "Redis URI")
	flag.StringVar(&config.cache.password, "redis_password", os.Getenv("REDIS_PASSWORD"), "Redis Password")
	REDIS_INDEX, err := strconv.Atoi(os.Getenv("REDIS_INDEX"))
	if err != nil {
		log.Fatal("failed setting redis index", err)
	}
	flag.IntVar(&config.cache.index, "redis_index", REDIS_INDEX, "Redis Cache Number")

	//?SMTP
	flag.StringVar(&config.smtp.host, "smtp_host", os.Getenv("SMTP_HOST"), "SMTP host")
	SMTP_PORT, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	flag.IntVar(&config.smtp.port, "smtp_port", SMTP_PORT, "SMTP port")
	flag.StringVar(&config.smtp.username, "smtp_username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&config.smtp.password, "smtp_password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&config.smtp.sender, "smtp_sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	flag.Parse()
	log.Info("command line variables loaded")

	postgres_client, err := initializePostgresClient(config)
	if err != nil {
		log.Fatal("failed initializing postgres client", err)
	}
	log.Info("postgres client initialized")

	redis_client, err := initializeRedisClient(config)
	if err != nil {
		log.Fatal("failed initializing redis client", err)
	}
	log.Info("redis client initialized")

	s3_client, err := initializeS3Client()
	if err != nil {
		log.Fatal("failed initializing s3 client", err)
	}
	log.Info("s3 client initialized")

	mail_client, err := initializeMailClient(config)
	if err != nil {
		log.Fatal("failed initializing mail client", err)
	}
	log.Info("mail client initialized")

	return mail_client, s3_client, postgres_client, redis_client, config.port
}
