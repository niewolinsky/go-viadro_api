package main

import (
	"viadro_api/config"
	"viadro_api/internal/data"
	"viadro_api/internal/logger"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/wneessen/go-mail"
)

type application struct {
	config       config.Config
	data_access  data.Layers
	s3_client    *s3.Client
	cache_client *redis.Client
	mail_client  *mail.Client
}

// @title						Viadro API
// @version					1.0.0
// @description				Open-source document hosting solution based on S3 storage.
// @contact.name				Viadro API Developer - Przemyslaw Niewolinski
// @contact.url				https://www.niewolinsky.dev
// @contact.email				niewolinski@protonmail.com
// @host						viadro.xyz:4000
// @BasePath					/v1/
// @securityDefinitions.basic	BasicAuth
// @schemes					https
// @produce					json
// @accept						json
// @accept						mpfd
// @license.name				MIT License
// @license.url				https://github.com/niewolinsky/go-viadro_api/blob/main/license.txt
func main() {
	mail_client, aws_s3_client, db_postgre, cache_client, cfg := config.InitConfig()
	defer db_postgre.Close()
	defer cache_client.Close()

	app := &application{
		config:       cfg,
		data_access:  data.NewLayers(db_postgre),
		s3_client:    aws_s3_client,
		cache_client: cache_client,
		mail_client:  mail_client,
	}

	err := app.serve()
	if err != nil {
		logger.LogFatal("failed starting server", err)
	}
}
