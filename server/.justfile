set dotenv-load := true

migrate:
    go run cmd/migrator/main.go

sqlc-gen:
    sqlc generate

serv:
    go run cmd/api/main.go

air:
    air -c air.toml
