package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
)


const (
	//DbUser     = "docker"
	//DbPassword = "docker"
	//DbName     = "docker"
	DbUser     = "tpforumsapi"
	DbPassword = "222"
	DbName = "forums_func"
	//DbName = "forums"
)

var db *sql.DB

func InitDb() (*sql.DB, error) {
	var err error
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DbUser, DbPassword, DbName)
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	init, err := ioutil.ReadFile("db/forum.sql")
	_, err = db.Exec(string(init))

	if err != nil {
		panic(err)
	}

	fmt.Println("You connected to your database.")

	return db, nil
}

func DbQueryRow(query string, args []interface{}) (*sql.Row){
	var row *sql.Row

	row = db.QueryRow(query, args...)

	return row

}

func DbQuery(query string, args []interface{}) (*sql.Rows, error) {
	var err error

	rows, err := db.Query(query, args...)

	return rows, err
}

func DbExec(query string, args []interface{}) (error) {
	//s := make([]interface{}, len(args))
	//for i, v := range args {
	//	s[i] = v
	//}
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


func GetLink()(*sql.DB){
	return db
}