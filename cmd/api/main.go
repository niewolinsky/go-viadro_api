package main

import (
	"sync"
	"viadro_api/config"
	"viadro_api/internal/data"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
	"github.com/wneessen/go-mail"
)

type application struct {
	data_access  data.Layers
	s3_client    *s3.Client
	redis_client *redis.Client
	mail_client  *mail.Client
	wait_group   sync.WaitGroup
}

// @title						Viadro API
// @version					0.7.0
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
	mail_client, s3_client, postgres_client, redis_client, app_port := config.InitConfig()
	defer postgres_client.Close()
	defer redis_client.Close()

	app := &application{
		data_access:  data.NewLayers(postgres_client),
		s3_client:    s3_client,
		redis_client: redis_client,
		mail_client:  mail_client,
	}

	err := app.serve(app_port)
	if err != nil {
		log.Fatal("failed starting server", err)
	}
	log.Info("stopped server")
}
