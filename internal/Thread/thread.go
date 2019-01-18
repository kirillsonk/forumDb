package Thread

import (
	"database/sql"
	// "forumDb/db"
	// "forumDb/internal/Errors"
	// "forumDb/models"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/kirillsonk/forumDb/Errors"
	"github.com/kirillsonk/forumDb/db"
	"github.com/kirillsonk/forumDb/models"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func PostsThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]
	Thread, err := GetThreadByIdOrSlug(slugOrId)
	w.Header().Set("content-type", "application/json")

	if err != nil {
		Errors.SendError("Can't find thread with id "+slugOrId, http.StatusNotFound, &w)
		return
	}

	descValue := r.URL.Query().Get("desc")
	sortValue := r.URL.Query().Get("sort")
	limitValue := r.URL.Query().Get("limit")
	sinceValue := r.URL.Query().Get("since")

	var dsc = false
	var lim = false
	var since = false

	if sinceValue != "" {
		since = true
	}
	if descValue == "true" {
		dsc = true
	}
	if limitValue == "" {
		limitValue = " ALL"
	} else {
		lim = true
	}
	if sortValue != "flat" && sortValue != "tree" && sortValue != "parent_tree" {
		sortValue = "flat"
	}

	var rows *sql.Rows

	if sortValue == "flat" {
		if dsc {
			if since {
				rows, err = db.DbQuery("SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE thread = $1 AND id < $3 ORDER BY created DESC, id DESC LIMIT $2", []interface{}{Thread.Id, limitValue, sinceValue})
			} else {
				rows, err = db.DbQuery("SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE thread = $1 ORDER BY id DESC LIMIT $2", []interface{}{Thread.Id, limitValue})
			}
		} else {
			if since {
				rows, err = db.DbQuery("SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE thread = $1 AND id > $3 ORDER BY id ASC LIMIT $2", []interface{}{Thread.Id, limitValue, sinceValue})
			} else {
				query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE thread = $1 ORDER BY id ASC LIMIT " + limitValue
				rows, err = db.DbQuery(query, []interface{}{Thread.Id})
			}
		}
	} else if sortValue == "parent_tree" {
		descFlag := ""
		sinceAdd := ""
		sortAdd := ""
		limitAdd := ""

		if lim != false {
			limitAdd = " WHERE rank <= " + limitValue
		}

		if dsc == true {
			descFlag = " desc "
			sortAdd = "ORDER BY id_array[1] DESC, id_array "
			if since != false {
				sinceAdd = " AND id_array[1] < (SELECT id_array[1] FROM Post WHERE id = " + sinceValue + " ) "
			}
		} else {
			descFlag = " ASC "
			sortAdd = " ORDER BY id_array[1], id_array ASC"
			if since != false {
				sinceAdd = " AND id_array[1] > (SELECT id_array[1] FROM Post WHERE id = " + sinceValue + " ) "
			}
		}

		query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM (" +
			" SELECT author,id_array,created,forum,id,isedited,message,parent,thread, " +
			" dense_rank() over (ORDER BY id_array[1] " + descFlag + " ) AS rank " +
			" FROM Post WHERE thread=$1 " + sinceAdd + " ) AS tree " + limitAdd + " " + sortAdd

		rows, err = db.DbQuery(query, []interface{}{Thread.Id})
	} else if sortValue == "tree" {
		sinceAdd := ""
		sortAdd := ""
		limitAdd := ""

		if lim != false {
			limitAdd = "LIMIT " + limitValue
		}

		if dsc == true {
			sortAdd = " ORDER BY id_array[0], id_array DESC "
			if since != false {
				sinceAdd = " AND id_array < (SELECT id_array FROM Post WHERE id = " + sinceValue + " ) "
			}
		} else {
			sortAdd = " ORDER BY id_array[0],id_array "
			if since != false {
				sinceAdd = " AND id_array > (SELECT id_array FROM Post WHERE id = " + sinceValue + " ) "
			}
		}

		query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE thread=$1 " + sinceAdd + " " + sortAdd + " " + limitAdd
		rows, err = db.DbQuery(query, []interface{}{Thread.Id})
	}

	if err != nil {
		// fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var i = 0
	postList := models.PostList{}
	for rows.Next() {
		i++
		post := models.Post{}

		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			// fmt.Println(err.Error())
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
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]
	w.Header().Set("content-type", "application/json")

	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		Thread := models.Thread{}
		err = Thread.UnmarshalJSON(body)

		if Thread.Title == "" && Thread.Message == "" {
			existThread, err := GetThreadByIdOrSlug(slugOrId)

			if err != nil {
				Errors.SendError("Can't find Thread with id "+slugOrId, http.StatusNotFound, &w)
				return
			}
			resData, _ := existThread.MarshalJSON()
			w.Write(resData)
			return
		}

		var add = " "
		var txtAdd = ""
		var titleAdd = ""

		if Thread.Message != "" {
			txtAdd = " message='" + Thread.Message + "' "
		}

		if Thread.Title != "" {
			titleAdd = " title='" + Thread.Title + "' "
		}

		if Thread.Title != "" && Thread.Message != "" {
			add = ","
		}

		var row *sql.Row

		thrId, err := strconv.Atoi(slugOrId)

		var idAdd string

		if err != nil {
			idAdd = "slug='" + slugOrId + "' "
		} else {
			idAdd = "id=" + strconv.Itoa(thrId)
		}

		data := "UPDATE Thread SET " + txtAdd + add + titleAdd + " WHERE " + idAdd + " RETURNING *"
		row = db.DbQueryRow(data, nil)

		err = row.Scan(&Thread.Id, &Thread.Author, &Thread.Created, &Thread.Forum, &Thread.Message, &Thread.Slug, &Thread.Title, &Thread.Votes)

		if err != nil {
			Errors.SendError("Can't find Thread with id "+slugOrId, http.StatusNotFound, &w)
			return
		}
		resData, err := Thread.MarshalJSON()
		w.Write(resData)
		return
	}
	Thread, err := GetThreadByIdOrSlug(slugOrId)

	if err != nil {
		Errors.SendError("Can't find Thread with id "+slugOrId, http.StatusNotFound, &w)
		return
	}

	resData, _ := Thread.MarshalJSON()
	w.Write(resData)
	return
}

func GetThreadByIdOrSlug(slug string) (*models.Thread, error) {
	thrId, err := strconv.Atoi(slug)
	var row *sql.Row

	if err != nil {
		row = db.DbQueryRow("SELECT * FROM Thread WHERE slug=$1;", []interface{}{slug})
	} else {
		row = db.DbQueryRow("SELECT * FROM Thread WHERE id=$1;", []interface{}{thrId})
	}

	var sqlSlug sql.NullString

	Thread := new(models.Thread)
	err = row.Scan(&Thread.Id, &Thread.Author, &Thread.Created, &Thread.Forum, &Thread.Message, &sqlSlug, &Thread.Title, &Thread.Votes)

	if !sqlSlug.Valid {
		Thread.Slug = ""
	} else {
		Thread.Slug = sqlSlug.String
	}

	if err != nil {
		return nil, err
	}

	return Thread, nil
}

func ThreadCreate(w http.ResponseWriter, r *http.Request) {
	body, readError := ioutil.ReadAll(r.Body)
	w.Header().Set("content-type", "application/json")
	defer r.Body.Close()

	if readError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var err error

	dbConn := db.GetLink()
	dbc, err := dbConn.Begin()

	if err != nil {
		// fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer dbc.Rollback()

	_, err = dbc.Exec("SET LOCAL synchronous_commit = OFF")

	if err != nil {
		// fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	Thread := models.Thread{}
	Thread.UnmarshalJSON(body)

	params := mux.Vars(r)
	slug := params["slug"]

	var row *sql.Row
	if Thread.Slug == "" {
		row = dbc.QueryRow("INSERT INTO Thread(author, created, forum, message, title) VALUES ($1, $2, "+
			"(SELECT slug FROM Forum WHERE slug=$3), $4, $5) RETURNING *", Thread.Author, Thread.Created, slug,
			Thread.Message, Thread.Title)
	} else {
		row = dbc.QueryRow("INSERT INTO Thread(author, created, forum, message, title, slug) VALUES ($1, $2, "+
			"(SELECT slug FROM Forum WHERE slug=$3), $4, $5, $6) RETURNING *", Thread.Author, Thread.Created, slug,
			Thread.Message, Thread.Title, Thread.Slug)
	}

	addedThread := models.Thread{}
	var sqlSlug sql.NullString
	err = row.Scan(&addedThread.Id, &addedThread.Author, &addedThread.Created, &addedThread.Forum, &addedThread.Message, &sqlSlug, &addedThread.Title, &addedThread.Votes)

	if err != nil {
		// fmt.Println(err.Error())
		errorName := err.(*pq.Error).Code.Name()

		if errorName == "foreign_key_violation" || errorName == "not_null_violation" {
			Errors.SendError("Can't find forum or usr", http.StatusNotFound, &w)
			return
		}

		if errorName == "unique_violation" {
			existThread, _ := GetThreadByIdOrSlug(Thread.Slug)

			w.WriteHeader(http.StatusConflict)
			resData, _ := existThread.MarshalJSON()
			w.Write(resData)
			return
		}
		return
	}

	_, err = dbc.Exec("INSERT INTO ForumUser(forum,author) VALUES ($1,$2) ON CONFLICT DO NOTHING", slug, Thread.Author)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !sqlSlug.Valid {
		addedThread.Slug = ""
	} else {
		addedThread.Slug = sqlSlug.String
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbc.Commit()

	resData, _ := addedThread.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(resData)
	return
}
