# Viadro - Cloud-based PDF document manager - API
Viadro is a cloud-native HTTP API for managing PDF documents in the cloud. It utilizes S3 storage under the hood, in a sense it is a light S3 wrapper for basic CRUD operations on files. The application works on it's own as an API, but for best user experience, to interact with a service, download: [Viadro CLI](https://github.com/niewolinsky/go-viadro_cli/). The service can be fully self-hosted making it an simple alternative to Dropbox/Google Drive if you are transfering a lot of documents and willing to sacrifice some advanced features.

![Open API docs](https://i.imgur.com/eES8vtu.png)

## Features:
- Token-based authentication system
- Upload, manage and delete documents (some features only work for PDF documents but you can also upload: .txt, .rtf, .docx and .md files)
- Access the documents from anywhere
- Get easily shareable link or hide your document from public repository
- Search through public repository of documents (with pagination and filters for: title, tag and document owner)
- Admin routes for advanced user and document management

### Additional features when using Viadro CLI:
- Dynamically search through list of public or user's private documents
- Merge many PDFs and upload with single command
- Grab a PDF from the web and host it on Viadro service instead
- Other UX improvements (autocompletions, help, status messages)

## Running:
Service is avaiable as a cloud solution for small group of alpha testers.

Meanwhile you can self-host the service

### Requirements
- PostgreSQL
- Redis
- SMTP provider
- Configured S3 bucket (access to secret keys for application)
- [Migrate](https://github.com/golang-migrate/migrate) tool

After configuring required services run database migrations using *migrate* tool, create `.env` file in project's root directory with KEY=VALUE:
<details>
  <summary>.env file example</summary>
  
      #APP ENV
      APP_PORT=
      APP_VERSION=
      APP_ENVIRONMENT=

      #AWS ENV
      AWS_ACCESS_KEY=
      AWS_SECRET_ACCESS_KEY=
      AWS_REGION=
      #AWS S3 ENV
      AWS_S3_BUCKET_NAME=

      #SMTP ENV
      SMTP_HOST=
      SMTP_SENDER=
      SMTP_PORT=
      SMTP_USERNAME=
      SMTP_PASSWORD=

      #POSTGRES ENV
      POSTGRES_DSN=

      #REDIS ENV
      REDIS_DSN=
      REDIS_PASSWORD=
      REDIS_INDEX=
</details>

## Todo:
- Password reset feature
- Remove user's files on account deletion
- User input validation
- Add owner's username to list of documents response
- File encryption

## Stack:
- Go 1.20 + [valyala/fasthttp](https://github.com/valyala/fasthttp) + [charmbracelet/log](https://github.com/charmbracelet/log) + [jackc/pgx](https://github.com/jackc/pgx) + [aws/aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) + [swaggo/swag](https://github.com/swaggo/swag) + [redis/go-redis](https://github.com/redis/go-redis) + [wneessen/go-mail](github.com/wneessen/go-mail) + [joho/godotenv](github.com/joho/godotenv)
- PostgreSQL 14+
- Redis 6+
- [Migrate](https://github.com/golang-migrate/migrate)

### Service utilizes AWS cloud services:
- EC2, SES, S3, RDS, ElastiCache

![App Architecture](https://i.imgur.com/fxKVSEH.png)

## Additional info
Application is actively developed and it is not production ready yet.
