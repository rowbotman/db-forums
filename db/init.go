package db

import (
	//"database/sql"
	//"encoding/json"
	//"fmt"
	"gopkg.in/jackc/pgx.v2"
	//"github.com/jackc/pgx"
	//"github.com/lib/pq"
	//"gopkg.in/jackc/pgx.v3"
	//"reflect"
	//"time"
)

type NullInt64 pgx.NullInt64

type NullBool pgx.NullBool

type NullFloat64 pgx.NullFloat64

type NullTime pgx.NullTime

type NullString pgx.NullString

var DB *pgx.ConnPool

//func (ni *NullInt64) Scan(value interface{}) error {
//	var i pgx.NullInt64
//	if err := i.Scan(pgx.ValueReader{}); err != nil {
//		return err
//	}
//	// if nil the make Valid false
//	if reflect.TypeOf(value) == nil {
//		*ni = NullInt64{Int64: i.Int64}
//	} else {
//		*ni = NullInt64{Int64: i.Int64, Valid: true}
//	}
//	return nil
//}
//
//func (ns *NullString) Scan(value interface{}) error {
//	var s pgx.NullString
//	if err := s.Scan(value); err != nil {
//		return err
//	}
//
//	// if nil then make Valid false
//	if reflect.TypeOf(value) == nil {
//		*ns = NullString{String: s.String}
//	} else {
//		*ns = NullString{String: s.String, Valid: true}
//	}
//
//	return nil
//}
//
//func (nt *NullTime) Scan(value interface{}) error {
//	var t pgx.NullTime
//	if err := t.Scan(value); err != nil {
//		return err
//	}
//
//	// if nil then make Valid false
//	if reflect.TypeOf(value) == nil {
//		*nt = NullTime{Time: t.Time}
//	} else {
//		*nt = NullTime{Time: t.Time, Valid: true}
//	}
//
//	return nil
//}
//
//func (ni *NullInt64) MarshalJSON() ([]byte, error) {
//	if !ni.Valid {
//		return []byte("null"), nil
//	}
//	return json.Marshal(ni.Int64)
//}
//
//func (ni *NullInt64) UnmarshalJSON(b []byte) error {
//	err := json.Unmarshal(b, &ni.Int64)
//	ni.Valid = err == nil
//	return err
//}
//
//func (ns *NullString) MarshalJSON() ([]byte, error) {
//	if !ns.Valid {
//		return []byte("null"), nil
//	}
//	return json.Marshal(ns.String)
//}
//
//// UnmarshalJSON for NullString
//func (ns *NullString) UnmarshalJSON(b []byte) error {
//	err := json.Unmarshal(b, &ns.String)
//	ns.Valid = err == nil
//	return err
//}
//
//func (nt *NullTime) MarshalJSON() ([]byte, error) {
//	if !nt.Valid {
//		return []byte("null"), nil
//	}
//	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
//	return []byte(val), nil
//}
//
//// UnmarshalJSON for NullTime
//func (nt *NullTime) UnmarshalJSON(b []byte) error {
//	s := string(b)
//	// s = Stripchars(s, "\"")
//
//	x, err := time.Parse(time.RFC3339, s)
//	if err != nil {
//		nt.Valid = false
//		return err
//	}
//
//	nt.Time = x
//	nt.Valid = true
//	return nil
//}

//func InitDB(dataSourceName string) *sql.DB {
//	var err error
//	DB, err = sql.Open("postgres", dataSourceName)
//	if err != nil {
//		log.Panic(err)
//	}
//
//	if err = DB.Ping(); err != nil {
//		log.Panic(err)
//	}
//	return DB
//}