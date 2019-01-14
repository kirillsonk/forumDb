package Forum

import (
	"ForumsApi/db"
	"ForumsApi/internal/Errors"
	"ForumsApi/internal/User"
	"ForumsApi/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"io/ioutil"
	"net/http"
)

func ForumUsers(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		return
	}

	limitVal := r.URL.Query().Get("limit")
	sinceVal := r.URL.Query().Get("since")
	descVal := r.URL.Query().Get("desc")

	var limit = false
	var since = false
	var desc = false

	if limitVal != "" {
		limit = true
	}
	if sinceVal != "" {
		since = true
	}
	if descVal == "true" {
		desc = true
	}

	var rows *sql.Rows

	var err error

	vars := mux.Vars(r)
	slug := vars["slug"]

	frm, _ := getForum(slug, nil)

	if frm == nil {
		Errors.SendError("Can't find forum with slug " + slug + "\n", 404, &w)
		return
	}


	if !limit && !since && !desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 ORDER BY nickname ASC"

		rows, err = db.DbQuery(query, []interface{}{slug})
	} else if !limit && !since && desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 ORDER BY nickname DESC "

		rows, err = db.DbQuery(query, []interface{}{slug})
	} else if !limit && since && !desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 AND u.nickname>$2 ORDER BY nickname ASC"

		rows, err = db.DbQuery(query, []interface{}{slug, sinceVal})

	} else if !limit && since && desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 AND u.nickname<$2 ORDER BY nickname DESC "
		rows, err = db.DbQuery(query, []interface{}{slug, sinceVal})

	} else if limit && !since && !desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 ORDER BY nickname ASC LIMIT $2"
		rows, err = db.DbQuery(query, []interface{}{slug, limitVal})

	} else if limit && !since && desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 ORDER BY nickname DESC LIMIT $2"
		rows, err = db.DbQuery(query, []interface{}{slug, limitVal})

	} else if limit && since && !desc {//here
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 AND u.nickname>$2 ORDER BY nickname ASC LIMIT $3"

		rows, err = db.DbQuery(query, []interface{}{slug, sinceVal, limitVal})

	} else if limit && since && desc {
		query := "SELECT about,email,fullname,nickname FROM forum_users f_u JOIN users u ON f_u.author=u.nickname AND f_u.forum=$1 AND u.nickname<$2 ORDER BY nickname DESC LIMIT $3"

		rows, err = db.DbQuery(query, []interface{}{slug, sinceVal, limitVal})

	}

	if err != nil {
		Errors.SendError( "Can't find forum with slug " + slug + "\n", 404, &w)
		return
	}

	users := make([]models.User, 0)

	for rows.Next() {
		usr := models.User{}

		err := rows.Scan(&usr.About, &usr.Email, &usr.FullName, &usr.NickName)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		users = append(users, usr)
	}

	defer rows.Close()


	resp, _ := json.Marshal(users)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)

	return
}


func ForumThreads(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		return
	}

	limitVal := r.URL.Query().Get("limit")
	sinceVal := r.URL.Query().Get("since")
	descVal := r.URL.Query().Get("desc")

	var limit = false
	var since = false
	var desc = false

	if limitVal != "" {
		limit = true
	}
	if sinceVal != "" {
		since = true
	}
	if descVal == "true" {
		desc = true
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	//frm, _ := getForum(slug, nil)  // Исправить
	//if frm == nil {
	//	Errors.SendError()("Can't find forum with slug " + slug + "\n", 404, &w)
	//	return
	//}

	var rows *sql.Rows

	var err error

	if limit && !since && !desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 ORDER BY created LIMIT $2;", []interface{}{slug, limitVal})
	} else if since && !limit && !desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created;", []interface{}{slug, sinceVal})
	} else if limit && since && !desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", []interface{}{slug, sinceVal, limitVal})
	} else if limit && !since && desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2;", []interface{}{slug, limitVal})
	} else if since && !limit && desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC;", []interface{}{slug, sinceVal})
	} else if limit && since && desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3;", []interface{}{slug, sinceVal, limitVal})
	} else if limit && since && !desc{
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", []interface{}{slug, sinceVal, limitVal})
	} else if !limit && !since && !desc {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 ORDER BY created;", []interface{}{slug})
	} else {
		rows, err = db.DbQuery("SELECT * FROM threads WHERE forum = $1 ORDER BY created;", []interface{}{slug})
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	thrs := make([]models.Thread, 0)

	var nullSlug sql.NullString

	var flag = false

	for rows.Next() {
		flag = true
		thr := models.Thread{}
		err := rows.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &nullSlug, &thr.Title, &thr.Votes)

		if nullSlug.Valid {
			thr.Slug = nullSlug.String
		} else {
			thr.Slug = ""
		}

		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		thrs = append(thrs, thr)
	}

	if flag == false {
		frm, _ := getForum(slug, nil)
		if frm == nil {
			Errors.SendError("Can't find forum with slug " + slug + "\n", 404, &w)
			return
		}
	}

	defer rows.Close()


	resp, _ := json.Marshal(thrs)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)

	return
}


func getForum(slugOrId string, t *sql.Tx) (*models.Forum,error) {
	forum := models.Forum{}
	var err error
	//if t == nil {
	err = db.DbQueryRow("SELECT * FROM forums WHERE slug=$1", []interface{}{slugOrId}).Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
	//} else {
	//	err = t.QueryRow("SELECT * FROM forums WHERE slug=$1", slugOrId).Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
	//}


	if err != nil {
		return nil, err
	}

	return &forum, nil
}


func ForumDetails(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	slug := vars["slug"]
	frm, err := getForum(slug, nil)

	if err != nil {
		Errors.SendError( "Can't find forum with slug " + slug + "\n", 404, &w)
		return
	}

	resp, err := json.Marshal(frm)

	if err != nil {
		return
	}
	w.Header().Set("content-type", "application/json")

	w.Write(resp)
	return
}


func ForumCreate(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	dbConn := db.GetLink()

	t, err := dbConn.Begin()

	if err != nil {
		fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer t.Rollback()

	_, err = t.Exec("SET LOCAL synchronous_commit TO OFF")

	if err != nil {
		fmt.Println("set local ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	forum := new(models.Forum)
	err = json.Unmarshal(body, forum)

	existUser, _ := User.GetUser(forum.User)

	if existUser == nil {
		Errors.SendError( "Can't find user with name " + forum.User + "\n", 404, &w)
		return
	}

	row := t.QueryRow("INSERT INTO forums(slug, title, author) VALUES ($1, $2, $3) RETURNING *", []interface{}{forum.Slug, forum.Title, existUser.NickName}...)

	err = row.Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

	if err != nil {
		errorName := err.(*pq.Error).Code.Name()
		if errorName == "foreign_key_violation" {
			Errors.SendError( "Can't find user with name " + forum.User + "\n", 404, &w)
			return
		}
		if errorName == "unique_violation" {
			row := db.DbQueryRow("SELECT * FROM forums WHERE slug=$1", []interface{}{forum.Slug})
			fr := models.Forum{}
			err := row.Scan(&fr.Posts, &fr.Slug, &fr.Threads, &fr.Title, &fr.User)

			if err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusConflict)
			resp, _ := json.Marshal(fr)

			w.Write(resp)
			return
		}
	}

	t.Commit()

	resp, _ := json.Marshal(forum)

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)

	w.Write(resp)

	return
}
