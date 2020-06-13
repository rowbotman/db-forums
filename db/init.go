package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

var db *pgxpool.Pool

func init() {
	//pgxConfig, _ := pgx.ParseURI("postgres://role1:12345@localhost:5432/docker")
	config, err := pgxpool.ParseConfig("postgres://park:admin@localhost:5432/park_forum")
	if err != nil {
		log.Fatal(err.Error())
	}
	db, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err.Error())
	}
	//db, err = pgx.NewConnPool(
	//	pgx.ConnPoolConfig{
	//		ConnConfig: pgxConfig,
	//	})
	//if err != nil {
	//	panic(err)
	//}

	//initSql, err := ioutil.ReadFile("init.sql")
	//if err != nil {
	//	panic(err)
	//}
	//initString := string(initSql)
	//
	//_, err = db.Exec(initString)
	//if err != nil {
	//	panic(err)
	//}
}

func GetDB() *pgxpool.Pool {
	return db
}
