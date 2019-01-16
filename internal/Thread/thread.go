package Thread

import (
	"forum-database/db"
	"forum-database/internal/Errors"
	"forum-database/models"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func ThreadPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	thr, err := GetThread(slugOrId)

	if err != nil {
		Errors.SendError("Can't find thread with id "+slugOrId+"\n", 404, &w)
		return
	}

	limitVal := r.URL.Query().Get("limit")
	sinceVal := r.URL.Query().Get("since")
	descVal := r.URL.Query().Get("desc")
	sortVal := r.URL.Query().Get("sort")

	var since = false
	var desc = false
	var limit = false

	if limitVal == "" {
		limitVal = " ALL"
	} else {
		limit = true
	}
	if sinceVal != "" {
		since = true
	}
	if descVal == "true" {
		desc = true
	}
	if sortVal != "flat" && sortVal != "tree" && sortVal != "parent_tree" {
		sortVal = "flat"
	}

	var rows *sql.Rows

	if sortVal == "flat" {
		if desc {

			if since {

				rows, err = db.DbQuery("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE thread = $1 AND id < $3 ORDER BY created DESC, id DESC LIMIT $2", []interface{}{thr.Id, limitVal, sinceVal})

			} else {

				rows, err = db.DbQuery("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE thread = $1 ORDER BY id DESC LIMIT $2", []interface{}{thr.Id, limitVal})

			}

		} else {

			if since {

				rows, err = db.DbQuery("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE thread = $1 AND id > $3 ORDER BY id ASC LIMIT $2", []interface{}{thr.Id, limitVal, sinceVal})

			} else {
				query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE thread = $1 ORDER BY id ASC LIMIT " + limitVal
				rows, err = db.DbQuery(query, []interface{}{thr.Id})

			}

		}
	} else if sortVal == "tree" {
		sinceAddition := ""
		sortAddition := ""
		limitAddition := ""
		if desc == true {
			sortAddition = " ORDER BY id_array[0], id_array DESC "
			if since != false {
				sinceAddition = " AND id_array < (SELECT id_array FROM posts WHERE id = " + sinceVal + " ) "
			}
		} else {
			sortAddition = " ORDER BY id_array[0],id_array "
			if since != false {
				sinceAddition = " AND id_array > (SELECT id_array FROM posts WHERE id = " + sinceVal + " ) "
			}
		}

		if limit != false {
			limitAddition = "LIMIT " + limitVal
		}
		query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE thread=$1 " + sinceAddition + " " + sortAddition + " " + limitAddition
		rows, err = db.DbQuery(query, []interface{}{thr.Id})
	} else if sortVal == "parent_tree" {
		descflag := ""
		sinceAddition := ""
		sortAddition := ""
		limitAddition := ""
		if desc == true {
			descflag = " desc "
			sortAddition = "ORDER BY id_array[1] DESC, id_array "
			if since != false {
				sinceAddition = " AND id_array[1] < (SELECT id_array[1] FROM posts WHERE id = " + sinceVal + " ) "
			}
		} else {
			descflag = " ASC "
			sortAddition = " ORDER BY id_array[1], id_array ASC"
			if since != false {
				sinceAddition = " AND id_array[1] > (SELECT id_array[1] FROM posts WHERE id = " + sinceVal + " ) "
			}
		}

		if limit != false {
			limitAddition = " WHERE rank <= " + limitVal
		}

		query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM (" +
			" SELECT author,id_array,created,forum,id,isedited,message,parent,thread, " +
			" dense_rank() over (ORDER BY id_array[1] " + descflag + " ) AS rank " +
			" FROM posts WHERE thread=$1 " + sinceAddition + " ) AS tree " + limitAddition + " " + sortAddition

		rows, err = db.DbQuery(query, []interface{}{thr.Id})
	}

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	postList := models.PostList{}
	var i = 0
	for rows.Next() {
		i++
		post := models.Post{}

		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		postList = append(postList, post)

	}

	defer rows.Close()

	resData, _ := postList.MarshalJSON()
	w.Write(resData)
	return
}

func ThreadDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	if r.Method == http.MethodPost {

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		thr := models.Thread{}
		err = thr.UnmarshalJSON(body)

		if thr.Title == "" && thr.Message == "" {
			existThr, err := GetThread(slugOrId)

			if err != nil {
				Errors.SendError("Can't find thread with id "+slugOrId+"\n", 404, &w)
				return
			}

			resData, _ := existThr.MarshalJSON()
			w.Write(resData)
			return
		}

		var many = " "

		var messageAddition = ""

		var titleAddition = ""

		if thr.Message != "" {
			messageAddition = " message='" + thr.Message + "' "
		}

		if thr.Title != "" {
			titleAddition = " title='" + thr.Title + "' "
		}

		if thr.Title != "" && thr.Message != "" {
			many = ","
		}

		var row *sql.Row

		thrId, err := strconv.Atoi(slugOrId)

		var idenAdditional string

		if err != nil {
			idenAdditional = "slug='" + slugOrId + "' "

		} else {
			idenAdditional = "id=" + strconv.Itoa(thrId)
		}

		query := "UPDATE threads SET " + messageAddition + many + titleAddition + " WHERE " + idenAdditional + " RETURNING *"
		row = db.DbQueryRow(query, nil)

		err = row.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum, &thr.Message, &thr.Slug, &thr.Title, &thr.Votes)

		if err != nil {
			Errors.SendError("Can't find thread with id "+slugOrId+"\n", 404, &w)
			return
		}
		resData, err := thr.MarshalJSON()
		w.Write(resData)
		return
	}

	thr, err := GetThread(slugOrId)

	if err != nil {
		Errors.SendError("Can't find thread with id "+slugOrId+"\n", http.StatusNotFound, &w)
		return
	}

	resData, _ := thr.MarshalJSON()
	w.Write(resData)
	return
}

func GetThread(slug string) (*models.Thread, error) {
	thrId, err := strconv.Atoi(slug)
	var row *sql.Row

	if err != nil {
		row = db.DbQueryRow("SELECT * FROM threads WHERE slug=$1;", []interface{}{slug})
	} else {
		row = db.DbQueryRow("SELECT * FROM threads WHERE id=$1;", []interface{}{thrId})
	}

	var sqlSlug sql.NullString

	thr := new(models.Thread)
	err = row.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum, &thr.Message, &sqlSlug, &thr.Title, &thr.Votes)

	if !sqlSlug.Valid {
		thr.Slug = ""
	} else {
		thr.Slug = sqlSlug.String
	}

	if err != nil {
		return nil, err
	}

	return thr, nil
}

func ThreadCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	body, readErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if readErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var err error

	dbConn := db.GetLink()

	t, err := dbConn.Begin()

	if err != nil {
		fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer t.Rollback()

	_, err = t.Exec("SET LOCAL synchronous_commit = OFF")

	if err != nil {
		fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	thr := models.Thread{}
	thr.UnmarshalJSON(body)

	params := mux.Vars(r)
	slug := params["slug"]

	var row *sql.Row
	if thr.Slug == "" {
		row = t.QueryRow("INSERT INTO threads(author, created, forum, message, title) VALUES ($1, $2, "+
			"(SELECT slug FROM forums WHERE slug=$3), $4, $5) RETURNING *", thr.Author, thr.Created, slug,
			thr.Message, thr.Title)
	} else {
		row = t.QueryRow("INSERT INTO threads(author, created, forum, message, title, slug) VALUES ($1, $2, "+
			"(SELECT slug FROM forums WHERE slug=$3), $4, $5, $6) RETURNING *", thr.Author, thr.Created, slug,
			thr.Message, thr.Title, thr.Slug)
	}

	newThr := models.Thread{}
	var sqlSlug sql.NullString
	err = row.Scan(&newThr.Id, &newThr.Author, &newThr.Created, &newThr.Forum, &newThr.Message, &sqlSlug, &newThr.Title, &newThr.Votes)

	if err != nil {
		fmt.Println(err.Error())
		errorName := err.(*pq.Error).Code.Name()

		if errorName == "foreign_key_violation" || errorName == "not_null_violation" {
			Errors.SendError("Can't find user or forum \n", 404, &w)
			return
		}

		if errorName == "unique_violation" {
			existThr, _ := GetThread(thr.Slug)

			w.WriteHeader(http.StatusConflict)
			resData, _ := existThr.MarshalJSON()
			w.Write(resData)
			return
		}
		return
	}

	_, err = t.Exec("INSERT INTO forum_users(forum,author) VALUES ($1,$2) ON CONFLICT DO NOTHING", slug, thr.Author)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !sqlSlug.Valid {
		newThr.Slug = ""
	} else {
		newThr.Slug = sqlSlug.String
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.Commit()

	resData, _ := newThr.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(resData)
	return
}
