package main

import (
	"viadro_api/config"
	"viadro_api/internal/data"
	"viadro_api/internal/logger"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/wneessen/go-mail"
)

type application struct {
	config      config.Config
	data_access data.Layers
	s3_manager  *manager.Uploader
	mail_client *mail.Client
}

// @title           Viadro API
// @version         0.1.0
// @description     Open-source document hosting solution
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @host      localhost:4000
// @BasePath  /v1/
// @securityDefinitions.basic  BasicAuth
func main() {
	mail_client, aws_s3_manager, db_postgre, cfg := config.InitConfig()
	defer db_postgre.Close()

	app := &application{
		config:      cfg,
		data_access: data.NewLayers(db_postgre),
		s3_manager:  aws_s3_manager,
		mail_client: mail_client,
	}

	err := app.serve()
	if err != nil {
		logger.LogFatal("failed starting server", err)
	}
}
