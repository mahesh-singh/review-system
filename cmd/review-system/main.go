package main

import (
	"context"
	"database/sql"
	"flag"
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
		dsn string
	}
	aws struct {
		region    string
		accessKey string
		secretKey string
		s3bucket  string
	}
}

type application struct {
	config appConfig
	logger *slog.Logger
	models data.Models
}

func main() {

	var cfg appConfig

	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://review:pa55word@localhost:5432/review?sslmode=disable", "PostgreSQL DSN")

	flag.StringVar(&cfg.aws.region, "aws-region", "ap-southeast-2", "AWS Region")
	flag.StringVar(&cfg.aws.s3bucket, "s3-bucket", "zuzuhotelreview1", "S3 Bucket name")
	flag.StringVar(&cfg.aws.accessKey, "aws-access-key", "", "AWS Access Key")
	flag.StringVar(&cfg.aws.secretKey, "aws-secret-key", "", "AWS Secret Key")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(&cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("database connection tested")
	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	s3client, err := s3.NewClient(cfg.aws.region)

	if err != nil {
		app.logger.Error("error in connecting aws S3", err)
	}

	files, err := s3client.ListFiles(context.Background(), cfg.aws.s3bucket, "")
	if err != nil {
		app.logger.Error("error while listing the file")
		app.logger.Error(err.Error())
	}

	jsonl_processing_service := jsonl_processing.NewJSONLProcessingService(db, nil, logger)

	s3FileReader := s3.NewS3FileReader(s3client)

	processing_result, err := jsonl_processing_service.ProcessMultipleFiles(context.Background(), files, s3FileReader, 5)

	if err != nil {
		app.logger.Error("Error in processing json", slog.String("error", err.Error()))
	}

	app.logger.Info("Processing result", slog.Int("Processing Result Len", len(processing_result)))

	app.logger.Info("Everything look great")

}

func openDB(config *appConfig) (*sql.DB, error) {
	fmt.Println(config.db.dsn)
	db, err := sql.Open("postgres", config.db.dsn)

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
