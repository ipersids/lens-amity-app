package main

import (
	"database/sql"
	"lensamity/migrations"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")

	slog.Info(dbURL)

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	slog.Info("database successfully connected")

	goose.SetBaseFS(migrations.SQLfs)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "sql"); err != nil {
		panic(err)
	}

	slog.Info("database migrations finished successfully")
}
