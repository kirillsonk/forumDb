package Forum

import (
	"database/sql"
	"fmt"
	// "forumDb/db"
	// "forumDb/models"
	// "forumDb/packs/Errors"
	// "forumDb/packs/User"
	"io/ioutil"
	"net/http"

	"github.com/kirillsonk/forumDb/db"
	"github.com/kirillsonk/forumDb/models"
	"github.com/kirillsonk/forumDb/packs/Errors"
	"github.com/kirillsonk/forumDb/packs/User"

	"github.com/gorilla/mux"
	pgx "github.com/jackc/pgx"
)

func UsersForum(w http.ResponseWriter, r *http.Request) { //+
	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		limitValue := r.URL.Query().Get("limit")
		descValue := r.URL.Query().Get("desc")
		sinceValue := r.URL.Query().Get("since")

		var lim = false
		var dsc = false
		var since = false

		if limitValue != "" {
			lim = true
		}
		if sinceValue != "" {
			since = true
		}
		if descValue == "true" {
			dsc = true
		}

		var err error
		var rowsData *pgx.Rows

		vars := mux.Vars(r)
		slug := vars["slug"]

		ForumBySlug, _ := forumBySlugOrID(slug)

		if ForumBySlug == nil {
			Errors.SendError("Can't find forum with slug "+slug, http.StatusNotFound, &w)
			return
		}

		if !lim && !since && !dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 ORDER BY nickname ASC"

			rowsData, err = db.DbQuery(data, []interface{}{slug})
		} else if !lim && !since && dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 ORDER BY nickname DESC "

			rowsData, err = db.DbQuery(data, []interface{}{slug})
		} else if !lim && since && !dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 AND usr.nickname>$2 ORDER BY nickname ASC"

			rowsData, err = db.DbQuery(data, []interface{}{slug, sinceValue})

		} else if !lim && since && dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 AND usr.nickname<$2 ORDER BY nickname DESC "
			rowsData, err = db.DbQuery(data, []interface{}{slug, sinceValue})

		} else if lim && !since && !dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 ORDER BY nickname ASC LIMIT $2"
			rowsData, err = db.DbQuery(data, []interface{}{slug, limitValue})

		} else if lim && !since && dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 ORDER BY nickname DESC LIMIT $2"
			rowsData, err = db.DbQuery(data, []interface{}{slug, limitValue})

		} else if lim && since && !dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 AND usr.nickname>$2 ORDER BY nickname ASC LIMIT $3"

			rowsData, err = db.DbQuery(data, []interface{}{slug, sinceValue, limitValue})

		} else if lim && since && dsc {
			data := "SELECT about,email,fullname,nickname FROM ForumUser frm_usr JOIN Users usr ON frm_usr.author=usr.nickname AND frm_usr.forum=$1 AND usr.nickname<$2 ORDER BY nickname DESC LIMIT $3"

			rowsData, err = db.DbQuery(data, []interface{}{slug, sinceValue, limitValue})

		}

		if err != nil {
			Errors.SendError("Can't find forum with slug "+slug, http.StatusInternalServerError, &w)
			return
		}

		usrList := models.UserList{}

		for rowsData.Next() {
			user := models.User{}
			err := rowsData.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			usrList = append(usrList, user)
		}

		rowsData.Close()

		resData, _ := usrList.MarshalJSON() //правильно easyjson
		w.Write(resData)
		return
	}
	return
}

func ThreadsForum(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		limitValue := r.URL.Query().Get("limit")
		sinceValue := r.URL.Query().Get("since")
		descValue := r.URL.Query().Get("desc")

		var lim = false
		var since = false
		var dsc = false

		if limitValue != "" {
			lim = true
		}
		if sinceValue != "" {
			since = true
		}
		if descValue == "true" {
			dsc = true
		}

		vars := mux.Vars(r)
		slug := vars["slug"]

		var rowsData *pgx.Rows

		var err error

		if lim && !since && !dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 ORDER BY created LIMIT $2;", []interface{}{slug, limitValue})
		} else if since && !lim && !dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 AND created <= $2 ORDER BY created;", []interface{}{slug, sinceValue})
		} else if lim && since && !dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", []interface{}{slug, sinceValue, limitValue})
		} else if lim && !since && dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 ORDER BY created DESC LIMIT $2;", []interface{}{slug, limitValue})
		} else if since && !lim && dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 AND created <= $2 ORDER BY created DESC;", []interface{}{slug, sinceValue})
		} else if lim && since && dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3;", []interface{}{slug, sinceValue, limitValue})
		} else if lim && since && !dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", []interface{}{slug, sinceValue, limitValue})
		} else if !lim && !since && !dsc {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 ORDER BY created;", []interface{}{slug})
		} else {
			rowsData, err = db.DbQuery("SELECT * FROM Thread WHERE forum = $1 ORDER BY created;", []interface{}{slug})
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		threadList := models.ThreadList{}
		var nullSlug sql.NullString
		var flag = false

		for rowsData.Next() {
			flag = true
			thread := models.Thread{}
			err := rowsData.Scan(&thread.Id,
				&thread.Author,
				&thread.Created,
				&thread.Forum,
				&thread.Message,
				&nullSlug,
				&thread.Title,
				&thread.Votes)

			if nullSlug.Valid {
				thread.Slug = nullSlug.String
			} else {
				thread.Slug = ""
			}

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			threadList = append(threadList, thread) //easyjson
		}

		if flag == false {
			ForumBySlug, _ := forumBySlugOrID(slug)
			if ForumBySlug == nil {
				Errors.SendError("Can't find forum with slug "+slug, http.StatusNotFound, &w)
				return
			}
		}

		rowsData.Close()

		resData, _ := threadList.MarshalJSON()
		w.Write(resData)
		return
	}
	return

}

func forumBySlugOrID(slugOrId string) (*models.Forum, error) {
	Forum := models.Forum{}

	var err error
	err = db.DbQueryRow("SELECT * FROM Forum WHERE slug=$1", []interface{}{slugOrId}).Scan(&Forum.Posts, &Forum.Slug, &Forum.Threads, &Forum.Title, &Forum.User)

	if err != nil {
		return nil, err
	}
	return &Forum, nil
}

func ForumDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	slug := vars["slug"]
	Forum, err := forumBySlugOrID(slug)

	if err != nil {
		Errors.SendError("Can't find forum with slug "+slug, http.StatusNotFound, &w)
		return
	}

	resData, _ := Forum.MarshalJSON()
	w.Write(resData)
	return
}

func CreateForum(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		dbConn := db.GetLink()

		t, err := dbConn.Begin()

		if err != nil {
			// fmt.Println("db.begin ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer t.Rollback()

		_, err = t.Exec("SET LOCAL synchronous_commit TO OFF")

		if err != nil {
			// fmt.Println("set local ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		Forum := new(models.Forum)
		err = Forum.UnmarshalJSON(body)

		existUser, _ := User.GetUserByNick(Forum.User)

		if existUser == nil {
			Errors.SendError("Can't find user with name "+Forum.User, http.StatusNotFound, &w)
			return
		}

		row := t.QueryRow("INSERT INTO Forum(slug, title, author) VALUES ($1, $2, $3) RETURNING *", []interface{}{Forum.Slug, Forum.Title, existUser.Nickname}...)

		err = row.Scan(&Forum.Posts, &Forum.Slug, &Forum.Threads, &Forum.Title, &Forum.User)

		if err != nil {
			t.Rollback()
			fmt.Println(err.(pgx.PgError).Message)
			error1 := Errors.CheckDuplicateError("forum_slug_key")
			errorName := err.(pgx.PgError).Message
			if errorName == "" {
				Errors.SendError("Can't find user with name "+Forum.User, http.StatusNotFound, &w)
				return
			}
			// duplicate key value violates unique constraint "forum_slug_key"
			if errorName == error1 {
				row := db.DbQueryRow("SELECT * FROM Forum WHERE slug=$1", []interface{}{Forum.Slug})
				fr := models.Forum{}
				err := row.Scan(&fr.Posts, &fr.Slug, &fr.Threads, &fr.Title, &fr.User)

				if err != nil {
					// fmt.Println(err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusConflict)
				resData, _ := fr.MarshalJSON()
				w.Write(resData)
				return
			}
		}

		t.Commit()

		resData, _ := Forum.MarshalJSON()
		w.WriteHeader(http.StatusCreated)
		w.Write(resData)
		return
	}
	return
}
