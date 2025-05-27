# Review System

## Prerequisites 
- Docker
- Go 1.23.0

## Steps to Run
1. Configure AWS 
   1. Using machine default aws profile 
   2. modify .env file (copy example.env file as a .env)
      1. `BUCKET`
      2. `AWS_REGION`
      3. 
2. Run the DB container and run the DB migration 
   1. `make docker/up`

3. Run the application
   1. `make run/review` 


## Architecture 
- `cmd/review-system` entry point
- `internal/data` DB model
- `internal/s3` S3 client 
- `internal/service/jsonl_processing/service.go` Main login to import files 
  - `ProcessMultipleFiles` is an entry point 


## Note
- I have taken LLM help to generate few code and discuss few ideas. 
- Didn't added test, due to to lack of time 
- More detail validation can be added


