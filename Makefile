DB_NAME ?= "postgres"
DB_PORT ?= "5432"
DB_HOST ?= "localhost"
DB_USERNAME ?= "postgres"
DB_PASSWORD ?= "password"
DB_PARAMS ?= "sslmode=disable"
BCRYPT_SALT ?= "8"
JWT_SECRET ?= "secret"

.PHONY: run
run:
	DB_NAME=${DB_NAME} DB_PORT=${DB_PORT} DB_HOST=${DB_HOST} DB_USERNAME=${DB_USERNAME} 					\
	DB_PASSWORD=${DB_PASSWORD} DB_PARAMS=${DB_PARAMS} BCRYPT_SALT=${BCRYPT_SALT} JWT_SECRET=${JWT_SECRET}	\
	go run .

.PHONY: create-migration
create-migration:
	@migrate create -ext sql -dir migrations ${name}

.PHONY: migrate-up
migrate-up:
	@migrate -database ${host} -path migrations up

.PHONY: migrate-down
migrate-down:
	@migrate -database ${host} -path migrations down