-include .env
export


.PHONY: confirm
confirm:
	@echo -n 'Are you sure?[y/N]' && read ans && [ $${ans:N} = y ]

# ==================================================================================== # 
# DEVELOPMENT
# ==================================================================================== #

## db/migration/new name=$1: create a new database migration
.PHONY: db/migration/new
db/migration/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


.PHONY: db/sql
db/psql:
	psql ${DB_DSN_LOCAL}

## db/migration/up: apply all up database migrations...
.PHONY: db/migration/up
db/migration/up: confirm
	@echo 'Running migrations...'
	migrate -path ./migrations -database ${DB_DSN_LOCAL} up


## run/review: run the review system import 
.PHONY: run/review
run/review:
	@echo 'Running review system import...'
	go run ./cmd/review-system -db-dsn ${DB_DSN_LOCAL} -aws-region ${AWS_REGION} -s3-bucket ${BUCKET}


## docker/up: Start Docker container
.PHONY: docker/up
docker/up:
	@echo "Starting Docker containers..."
	docker compose up -d

## docker/down: Stop docker containers
.PHONY: docker/down
docker/down:
	@echo "Stopping Docker containers..."
	docker compose down

## docker/clean: Remove docker containers, network and volume
.PHONY: docker/clean
docker/clean:
	@echo "Removing Docker containers, networks, and volumes..."
	docker compose down -v

## docker/logs: Show logs of running containers
.PHONY: docker/logs
docker/logs:
	@echo "Show logs..."
	docker compose logs -f