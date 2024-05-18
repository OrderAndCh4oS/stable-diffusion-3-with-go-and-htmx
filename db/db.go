package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

var Bun *bun.DB

func CreateDatabase(
	dbname string,
	dbuser string,
	dbpassword string,
	dbhost string,
	dbport string,
) (*sql.DB, error) {
	uri := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbuser,
		dbpassword,
		dbname,
		dbhost,
		dbport,
	)
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Init() error {
	var (
		host   = os.Getenv("DB_HOST")
		user   = os.Getenv("DB_USER")
		pass   = os.Getenv("DB_PASSWORD")
		dbname = os.Getenv("DB_NAME")
		port   = os.Getenv("DB_PORT")
	)
	db, err := CreateDatabase(dbname, user, pass, host, port)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	Bun = bun.NewDB(db, pgdialect.New())
	if len(os.Getenv("APP_DEBUG")) > 0 {
		Bun.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}
	return nil
}
