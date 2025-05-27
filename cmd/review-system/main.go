package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mahesh-singh/review-system/internal/data"
	"github.com/mahesh-singh/review-system/internal/s3"
	"github.com/mahesh-singh/review-system/internal/service/jsonl_processing"
)

type appConfig struct {
	env string
	db  struct {
		dns string
	}
	aws struct {
		region    string
		accessKey string
		secretKey string
	}
}

type application struct {
	config appConfig
	logger *slog.Logger
	models data.Models
}

func main() {
	var config appConfig

	config.db.dns = "postgres://review:pa55word@localhost:5432/review?sslmode=disable"

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(&config)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("database connection tested")
	defer db.Close()

	app := &application{
		config: config,
		logger: logger,
		models: data.NewModels(db),
	}

	config.aws.region = "ap-southeast-2"

	s3client, err := s3.NewClient(config.aws.region)

	if err != nil {
		app.logger.Error("error in connecting aws S3", err)
	}

	files, err := s3client.ListFiles(context.Background(), "zuzuhotelreview1", "")
	if err != nil {
		app.logger.Error("error while listing the file")
		app.logger.Error(err.Error())
	}

	jsonl_processing_service := jsonl_processing.NewJSONLProcessingService(db, nil, logger)

	s3FileReader := s3.NewS3FileReader(s3client)

	for _, file := range files {
		reader, err := s3FileReader.GetReader(context.Background(), file.S3Path)
		if err != nil {
			app.logger.Error("error in reading the file", slog.String("s3path", file.S3Path))
			app.logger.Error(err.Error())
		}

		fmt.Print(reader)
	}

	processing_result, err := jsonl_processing_service.ProcessMultipleFiles(context.Background(), files, s3FileReader, 5)

	if err != nil {
		app.logger.Error("Error in processing json", err)
	}

	app.logger.Info("Processing result", slog.Int("Processing Result Len", len(processing_result)))

	app.logger.Info("Everything look great")

}

func openDB(config *appConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.db.dns)

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
