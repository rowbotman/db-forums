package db

import (
	"gopkg.in/jackc/pgx.v2"
)

type NullInt64 pgx.NullInt64

type NullBool pgx.NullBool

type NullFloat64 pgx.NullFloat64

type NullTime pgx.NullTime

type NullString pgx.NullString

var DB *pgx.ConnPool