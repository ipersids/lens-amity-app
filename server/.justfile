set dotenv-load := true
set fallback := true

migrate:
    go run cmd/migrator/main.go

sqlc-gen:
    sqlc generate

serv:
    go run cmd/api/main.go

air:
    air -c air.toml

goose-create NAME="NAME":
    goose create {{ NAME }} sql
