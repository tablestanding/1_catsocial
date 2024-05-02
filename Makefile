.EXPORT_ALL_VARIABLES:

DB_NAME ?= postgres
DB_PORT ?= 5432
DB_HOST ?= localhost
DB_USERNAME ?= postgres
DB_PASSWORD ?= password
DB_PARAMS ?= sslmode=disable
BCRYPT_SALT ?= 8
JWT_SECRET ?= secret
OTEL_RESOURCE_ATTRIBUTES ?= service.name=catsocial,service.version=0.0.1
OTEL_EXPORTER_OTLP_ENDPOINT ?= http://localhost:4317
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT ?= http://localhost:4317

.PHONY: run
run:
	@go run .

.PHONY: create-migration
create-migration:
	@migrate create -ext sql -dir migrations ${name}

.PHONY: migrate-up
migrate-up:
	@migrate -database ${host} -path migrations up

.PHONY: migrate-down
migrate-down:
	@migrate -database ${host} -path migrations down