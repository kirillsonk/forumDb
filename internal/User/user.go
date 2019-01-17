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

func GetUserByNick(nick string) (*models.User, error) {
	if nick == "" {
		return nil, nil
	}

	var qrRow *sql.Row

	qrRow = db.DbQueryRow("SELECT about,email,fullname,nickname FROM Users WHERE nickname=$1", []interface{}{nick})
	Usr := models.User{}
	err := qrRow.Scan(&Usr.About, &Usr.Email, &Usr.Fullname, &Usr.Nickname)

	if err != nil {
		return nil, err
	}

	return &Usr, nil
}

func UserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	usrNick := vars["nickname"]
	w.Header().Set("content-type", "application/json")

	if r.Method == http.MethodGet {
		Usr, err := GetUserByNick(usrNick)

		if err != nil {
			Errors.SendError("Can't find user with nickname "+usrNick, http.StatusNotFound, &w)
			return
		}
		resData, _ := Usr.MarshalJSON()
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
	fullName := false
	email := false

	if userUpdate.About != "" {
		about = true
	}
	if userUpdate.Fullname != "" {
		fullName = true
	}
	if userUpdate.Email != "" {
		email = true
	}

	if !email && !fullName && !about {
		Usr, err := GetUserByNick(usrNick)

		if err != nil {
			Errors.SendError("Can't find prifile with id "+usrNick, http.StatusNotFound, &w)
		}

		resData, _ := Usr.MarshalJSON()
		w.Write(resData)
		return
	}

	var qr string
	var qrRow *sql.Row

	if about && email && fullName {
		qr = "UPDATE Users SET about=$1, fullname=$2, email=$3 WHERE nickname=$4 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.About, userUpdate.Fullname, userUpdate.Email, usrNick})
	} else if about && !email && fullName {
		qr = "UPDATE Users SET about=$1, fullname=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.About, userUpdate.Fullname, usrNick})
	} else if about && !email && !fullName {
		qr = "UPDATE Users SET about=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.About, usrNick})
	} else if about && email && !fullName {
		qr = "UPDATE Users SET about=$1, email=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.About, userUpdate.Email, usrNick})
	} else if !about && email && fullName {
		qr = "UPDATE Users SET fullname=$1, email=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.Fullname, userUpdate.Email, usrNick})
	} else if !about && email && !fullName {
		qr = "UPDATE Users SET email=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.Email, usrNick})
	} else if !about && !email && fullName {
		qr = "UPDATE Users SET fullname=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		qrRow = db.DbQueryRow(qr, []interface{}{userUpdate.Fullname, usrNick})
	}

	err = qrRow.Scan(&userUpdate.About, &userUpdate.Email, &userUpdate.Fullname, &userUpdate.Nickname)

	if err != nil {
		if err == sql.ErrNoRows {
			Errors.SendError("Can't find prifile with id "+usrNick, http.StatusNotFound, &w)
			return
		}

		errorName := err.(*pq.Error).Code.Name()

		if errorName == "unique_violation" {
			Errors.SendError("Can't change prifile with id "+usrNick, http.StatusConflict, &w)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resData, _ := userUpdate.MarshalJSON()
	w.Write(resData)
	return
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	usrNick := vars["nickname"]
	body, _ := ioutil.ReadAll(r.Body)
	w.Header().Set("content-type", "application/json")
	defer r.Body.Close()

	Usr := models.User{}
	err := Usr.UnmarshalJSON(body)

	Usr.Nickname = usrNick

	if Usr.Nickname == "" || Usr.Email == "" || Usr.About == "" || Usr.Fullname == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dbConn := db.GetLink()
	dbc, err := dbConn.Begin()
	if err != nil {
		// fmt.Println("set local begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = dbc.Exec("SET LOCAL synchronous_commit TO OFF")

	if err != nil {
		// fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer dbc.Rollback()

	qr := "INSERT INTO Users(about, email, fullname, nickname) VALUES ($1,$2,$3,$4) RETURNING *"

	err = dbc.QueryRow(qr, Usr.About, Usr.Email, Usr.Fullname, Usr.Nickname).Scan(&Usr.About,
		&Usr.Email, &Usr.Fullname, &Usr.Nickname)

	if err != nil {
		// fmt.Println(err.Error())
		errorName := err.(*pq.Error).Code.Name()

		if errorName == "unique_violation" {
			usrList := models.UserList{}

			rows, err := db.DbQuery("SELECT * FROM Users WHERE nickname=$1 OR email=$2", []interface{}{Usr.Nickname, Usr.Email})

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for rows.Next() {
				usr := models.User{}
				err := rows.Scan(&usr.About, &usr.Email, &usr.Fullname, &usr.Nickname)

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
	resData, err := Usr.MarshalJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dbc.Commit()
	w.WriteHeader(http.StatusCreated)
	w.Write(resData)
	return
}
