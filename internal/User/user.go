package User

import (
	"database/sql"
	"forum-database/db"
	"forum-database/internal/Errors"
	"forum-database/models"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func GetUser(nickname string) (*models.User, error) {
	if nickname == "" {
		return nil, nil
	}

	var row *sql.Row

	row = db.DbQueryRow("SELECT about,email,fullname,nickname FROM users WHERE nickname=$1", []interface{}{nickname})

	user := models.User{}

	err := row.Scan(&user.About, &user.Email, &user.FullName, &user.NickName)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	if r.Method == http.MethodGet {
		user, err := GetUser(nickname)

		if err != nil {
			Errors.SendError("Can't find user with nickname "+nickname+"\n", 404, &w)
			return
		}

		resData, _ := user.MarshalJSON()
		w.Write(resData)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userUpdate := models.User{}
	err = userUpdate.UnmarshalJSON(body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	about := false
	fullname := false
	email := false

	if userUpdate.About != "" {
		about = true
	}
	if userUpdate.FullName != "" {
		fullname = true
	}
	if userUpdate.Email != "" {
		email = true
	}

	if !email && !fullname && !about {
		user, err := GetUser(nickname)

		if err != nil {
			Errors.SendError("Can't find prifile with id "+nickname+"\n", 404, &w)
		}

		resData, _ := user.MarshalJSON()
		w.Write(resData)
		return
	}

	var query string
	var row *sql.Row

	if about && fullname && email {
		query = "UPDATE users SET about=$1, fullname=$2, email=$3 WHERE nickname=$4 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.About, userUpdate.FullName, userUpdate.Email, nickname})
	} else if about && fullname && !email {
		query = "UPDATE users SET about=$1, fullname=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.About, userUpdate.FullName, nickname})
	} else if about && !fullname && email {
		query = "UPDATE users SET about=$1, email=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.About, userUpdate.Email, nickname})
	} else if about && !fullname && !email {
		query = "UPDATE users SET about=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.About, nickname})
	} else if !about && fullname && email {
		query = "UPDATE users SET fullname=$1, email=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.FullName, userUpdate.Email, nickname})
	} else if !about && fullname && !email {
		query = "UPDATE users SET fullname=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.FullName, nickname})
	} else if !about && !fullname && email {
		query = "UPDATE users SET email=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		row = db.DbQueryRow(query, []interface{}{userUpdate.Email, nickname})
	}

	err = row.Scan(&userUpdate.About, &userUpdate.Email, &userUpdate.FullName, &userUpdate.NickName)

	if err != nil {
		if err == sql.ErrNoRows {
			Errors.SendError("Can't find prifile with id "+nickname+"\n", 404, &w)
			return
		}

		errorName := err.(*pq.Error).Code.Name()

		if errorName == "unique_violation" {
			Errors.SendError("Can't change prifile with id "+nickname+"\n", 409, &w)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resData, _ := userUpdate.MarshalJSON()
	w.Write(resData)
	return
}

func UserCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	nickname := vars["nickname"]
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	user := models.User{}
	err := user.UnmarshalJSON(body)

	user.NickName = nickname

	if user.NickName == "" || user.About == "" || user.Email == "" || user.FullName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dbConn := db.GetLink()

	t, err := dbConn.Begin()

	if err != nil {
		// fmt.Println("set local begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = t.Exec("SET LOCAL synchronous_commit TO OFF")

	if err != nil {
		// fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer t.Rollback()

	query := "INSERT INTO users(about, email, fullname, nickname) VALUES ($1,$2,$3,$4) RETURNING *"

	err = t.QueryRow(query, user.About, user.Email, user.FullName, user.NickName).Scan(&user.About,
		&user.Email, &user.FullName, &user.NickName)

	if err != nil {
		// fmt.Println(err.Error())
		errorName := err.(*pq.Error).Code.Name()

		if errorName == "unique_violation" {
			usrList := models.UserList{}

			rows, err := db.DbQuery("SELECT * FROM users WHERE nickname=$1 OR email=$2", []interface{}{user.NickName, user.Email})

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for rows.Next() {
				usr := models.User{}

				err := rows.Scan(&usr.About, &usr.Email, &usr.FullName, &usr.NickName)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				usrList = append(usrList, usr)
			}
			rows.Close()

			resData, _ := usrList.MarshalJSON()
			w.WriteHeader(http.StatusConflict)
			w.Write(resData)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	resData, err := user.MarshalJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Commit()
	w.WriteHeader(http.StatusCreated)
	w.Write(resData)
	return

}
