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
	psql ${DB_DSN}

## db/migration/up: apply all up database migrations...
.PHONY: db/migration/up
db/migration/up: confirm
	@echo 'Running migrations...'
	migrate -path ./migrations -database ${DB_DSN} up
