.PHONY: create-migration
create-migration:
	@migrate create -ext sql -dir migrations ${name}

.PHONY: migrate-up
migrate-up:
	@migrate -database ${host} -path migrations up

.PHONY: migrate-down
migrate-down:
	@migrate -database ${host} -path migrations down