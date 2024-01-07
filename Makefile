migrate:
	go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations

run:
	go run ./cmd/sso --config=./config/config.yaml
