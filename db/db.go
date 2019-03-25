package db

import (
	"fmt"

	"database/sql"

	"github.com/jackc/pgx"
)

const (
	// user     = "Sonk"
	// password = "k123"
	// dbname   = "forumdb"
	user     = "postgres"
	password = "root"
	dbname   = "root"
)

// var db *pgx.Conn
var db *pgx.ConnPool
var dbsql *sql.DB

func InitDbSQL() (*sql.DB, error) {
	var err error
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		user, password, dbname)
	dbsql, err = sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}

	if err = dbsql.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("You connected to your database.")

	return dbsql, nil
}

func InitDatabase() (*pgx.ConnPool, error) {
	var err error
	// dbInfo := fmt.Sprintf("user=%s "+
	// 	"password=%s dbname=%s sslmode=disable",
	// 	user, password, dbname)
	dbInfo := pgx.ConnConfig{
		User:     user,
		Password: password,
		// Host:     "localhost",
		Host:     "35.228.7.66",
		Port:     5432,
		Database: dbname,
	}

	db, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     dbInfo,
		MaxConnections: 16,
	})

	if err != nil {
		panic(err)
	}

	// init, err := ioutil.ReadFile("db/tables.sql")
	// _, err = db.Exec(string(init))

	// if err != nil {
	// 	panic(err)
	// }

	fmt.Println("Connected to database")
	return db, nil
}

func DbQueryRow(query string, args []interface{}) *pgx.Row {
	var row *pgx.Row
	row = db.QueryRow(query, args...)
	return row
}

func DbQuery(query string, args []interface{}) (*pgx.Rows, error) {
	var err error
	rows, err := db.Query(query, args...)
	return rows, err
}

func DbExec(query string, args []interface{}) error {
	var err error
	t, err := db.Begin()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err = t.Exec("SET LOCAL synchronous_commit TO OFF")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	defer t.Rollback()

	if args != nil {
		_, err = t.Exec(query, args...)
	} else {
		_, err = t.Exec(query)
	}

	t.Commit()
	return err
}

func GetLink() *pgx.ConnPool {
	return db
}

func GetLinkSql() *sql.DB {
	return dbsql
}
