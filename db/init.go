package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

var db *pgxpool.Pool

func init() {
	config, err := pgxpool.ParseConfig("postgres://park:admin@localhost:5432/park_forum")
	if err != nil {
		log.Fatal(err.Error())
	}
	db, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetDB() *pgxpool.Pool {
	return db
}
