package Post

import (
	"database/sql"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx"

	// "forumDb/db"
	// "forumDb/models"
	// "forumDb/packs/Errors"
	// "forumDb/packs/Thread"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kirillsonk/forumDb/db"
	"github.com/kirillsonk/forumDb/models"
	"github.com/kirillsonk/forumDb/packs/Errors"
	"github.com/kirillsonk/forumDb/packs/Thread"

	"github.com/gorilla/mux"
)

// func CreatePost(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodPost {
// 		w.Header().Set("content-type", "application/json")
// 		vars := mux.Vars(r)
// 		IdorSlug := vars["slug_or_id"]

// 		body, err := ioutil.ReadAll(r.Body)
// 		defer r.Body.Close()

// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		postList := models.PostList{}
// 		err = postList.UnmarshalJSON(body)

// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		dbConnection := db.GetLink()
// 		dbc, err := dbConnection.Begin()

// 		if err != nil {
// 			// fmt.Println("db.begin ", err.Error())
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		_, err = dbc.Exec("SET LOCAL synchronous_commit TO OFF")

// 		if err != nil {
// 			// fmt.Println("set local ", err.Error())
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		Thread, err := Thread.GetThreadByIdOrSlug(IdorSlug)

// 		if err != nil {
// 			Errors.SendError("Can't find thread with id "+IdorSlug, http.StatusNotFound, &w)
// 			return
// 		}

// 		defer dbc.Rollback()
// 		if len(postList) == 0 {
// 			data := models.PostList{}
// 			resData, err := data.MarshalJSON()

// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}

// 			w.WriteHeader(http.StatusCreated)
// 			w.Write(resData)
// 			return
// 		}

// 		responseQuery := "INSERT INTO Post(author, forum, message, parent, thread) VALUES "

// 		UsersForumInsert := "INSERT INTO ForumUser(forum,author) VALUES "

// 		var subQuery []string
// 		var UsersForumSubQuery []string

// 		for _, Post := range postList {

// 			values := fmt.Sprintf("('%s', '%s', '%s', %d, %d) ", Post.Author, Thread.Forum, Post.Message, Post.Parent, Thread.Id)

// 			subQuery = append(subQuery, values)
// 			UsersForumValues := fmt.Sprintf("('%s', '%s') ", Thread.Forum, Post.Author)
// 			UsersForumSubQuery = append(UsersForumSubQuery, UsersForumValues)
// 		}

// 		responseQuery += strings.Join(subQuery, ",") + " RETURNING author,created,forum,id,isedited,message,parent,thread;"

// 		UsersForumInsert += strings.Join(UsersForumSubQuery, ",") + " ON CONFLICT DO NOTHING;"

// 		responseQuery += UsersForumInsert
// 		rows, err := dbc.QueryEx(responseQuery)

// 		if err != nil {
// 			dbc.Rollback()
// 			fmt.Println(err.(pgx.PgError).Message)
// 			errorName := err.(*pq.Error).Code.Name()

// 			// fmt.Println(errorName)
// 			if err.Error() == "pq: Parent Post exc" {
// 				Errors.SendError("Parent Post was created in another thread", http.StatusConflict, &w)
// 				return
// 			}

// 			if errorName == "foreign_key_violation" {
// 				Errors.SendError("Can't find parent Post", http.StatusNotFound, &w)
// 				return
// 			}

// 			if errorName != "syntax_error" {
// 				Errors.SendError("Can't find parent Post", http.StatusConflict, &w)
// 				return
// 			}

// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		// data := make([]models.Post, 0)
// 		data := models.PostList{}

// 		for rows.Next() {
// 			addedPost := models.Post{}

// 			err := rows.Scan(&addedPost.Author,
// 				&addedPost.Created,
// 				&addedPost.Forum,
// 				&addedPost.Id,
// 				&addedPost.IsEdited,
// 				&addedPost.Message,
// 				&addedPost.Parent,
// 				&addedPost.Thread)

// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}

// 			data = append(data, addedPost)
// 		}

// 		resData, err := data.MarshalJSON()
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		dbc.Commit()
// 		w.WriteHeader(http.StatusCreated)
// 		w.Write(resData)
// 		return
// 	}

// 	return
// }

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	posts := models.PostList{}
	err = posts.UnmarshalJSON(body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// data := make([]models.Post, 0)

	// t, err := db.GetLink()
	dbConnection := db.GetLink()
	t, err := dbConnection.Begin()

	if err != nil {
		fmt.Println("db.begin ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = t.Exec("SET LOCAL synchronous_commit TO OFF")

	if err != nil {
		fmt.Println("set local ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	thr, err := Thread.GetThreadByIdOrSlug(slugOrId)

	if err != nil {
		Errors.SendError("Can't find thread with id "+slugOrId, http.StatusNotFound, &w)
		return
	}
	//thr := new(models.Thread)
	//thrId, err := strconv.Atoi(slugOrId)
	//if err == nil {
	//	thr.Id = int32(thrId)
	//} else {
	//	thr, err = getThread(slugOrId)
	//
	//	if err != nil{
	//		Errors.sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
	//		return
	//	}
	//
	//}

	defer t.Rollback()

	var firstCreated time.Time
	var count = 0
	//var err error
	// _, err = t.Prepare("name", "INSERT INTO Post(author, forum, message, parent, thread, created) VALUES ($1,$2,$3,$4,$5,$6) RETURNING author,created,forum,id,isedited,message,parent,thread")

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	postList := models.PostList{}
	for _, p := range posts {
		newPost := models.Post{}

		if count == 0 { // Для того, чтобы все последующие добавления постов происхдили с той же датой и временем.
			row := t.QueryRow("INSERT INTO Post(author, forum, message, parent, thread) VALUES ($1,$2,$3,$4,$5) RETURNING author,created,forum,id,isedited,message,parent,thread",
				p.Author, thr.Forum, p.Message, p.Parent, thr.Id)
			err = row.Scan(&newPost.Author, &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
				&newPost.Parent, &newPost.Thread)

			firstCreated = newPost.Created
		} else {
			row := t.QueryRow("INSERT INTO Post(author, forum, message, parent, thread, created) VALUES ($1,$2,$3,$4,$5,$6) RETURNING author,created,forum,id,isedited,message,parent,thread", p.Author, thr.Forum, p.Message, p.Parent, thr.Id, firstCreated)
			err = row.Scan(&newPost.Author, &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
				&newPost.Parent, &newPost.Thread)
		}

		if err != nil {
			t.Rollback()
			fmt.Println(err.Error())
			errorName := err.(pgx.PgError).Message

			fmt.Println(errorName)
			if errorName == "Parent post exc" {
				Errors.SendError("Parent post was created in another thread", http.StatusConflict, &w)
				return
			}

			if errorName == "foreign_key_violation" {
				Errors.SendError("Can't find parent post", http.StatusNotFound, &w)
				return
			}

			Errors.SendError("Can't find parent post", http.StatusNotFound, &w)
			return
		}

		_, err := t.Exec("INSERT INTO ForumUser(forum,author) VALUES ($1,$2) ON CONFLICT DO NOTHING", thr.Forum, p.Author)

		if err != nil {
			fmt.Println("postCreate insert ForumUsers ", err.Error())
		}

		postList = append(postList, newPost)
		count++
	}

	resData, err := postList.MarshalJSON()
	// resp, err := json.Marshal(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Commit()

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)
	w.Write(resData)

	return

}

func PostDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]

	rel := r.URL.Query().Get("related")

	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		Post := new(models.Post)
		err = Post.UnmarshalJSON(body)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err != nil {
			Errors.SendError("Can't find Post with id "+id, http.StatusNotFound, &w)
			return
		}

		if Post.Message == "" {
			row := db.DbQueryRow("SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE id=$1", []interface{}{id})

			err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			resData, _ := Post.MarshalJSON()
			w.Write(resData)
			return
		}
		row := db.DbQueryRow("UPDATE Post SET message=$1, isedited=true WHERE id=$2 RETURNING author,created,forum,id,isedited,message,parent,thread", []interface{}{Post.Message, id})
		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread)

		if err != nil {
			Errors.SendError("Can't find Post with id "+id, http.StatusNotFound, &w)
			return
		}

		resData, _ := Post.MarshalJSON()
		w.Write(resData)
		return
	}

	postDetail := models.PostDetail{}
	var objArray []string
	objects := strings.Split(rel, ",")
	for index := range objects {
		itemData := objects[index]
		objArray = append(objArray, itemData)
	}

	var relatedFrm = false
	var relatedUsr = false
	var relatedThr = false

	for index := range objArray {
		if objArray[index] == "forum" {
			relatedFrm = true
		}
		if objArray[index] == "user" {
			relatedUsr = true
		}
		if objArray[index] == "thread" {
			relatedThr = true
		}
	}

	var err error
	var sqlSlug sql.NullString

	if !relatedThr && !relatedUsr && !relatedFrm {
		Post := new(models.Post)
		row := db.DbQueryRow("SELECT author,created,forum,id,isedited,message,parent,thread FROM Post WHERE id=$1;", []interface{}{id})

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread)

		postDetail.Post = Post

	} else if !relatedUsr && !relatedThr && relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, frm.posts, frm.slug, frm.threads, frm.title, frm.author FROM Post psts " +
			"JOIN Forum frm ON psts.id=$1 AND psts.forum=frm.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		Forum := new(models.Forum)
		Post := new(models.Post)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&Forum.Posts, &Forum.Slug, &Forum.Threads, &Forum.Title, &Forum.User)

		postDetail.Forum = Forum
		postDetail.Post = Post

	} else if !relatedUsr && relatedThr && !relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, thrs.id, thrs.author, thrs.created, thrs.forum, thrs.message, thrs.slug, thrs.title, thrs.votes FROM Post psts " +
			"JOIN Thread thrs ON psts.id=$1 AND psts.thread=thrs.id"
		row := db.DbQueryRow(query, []interface{}{id})

		Thread := new(models.Thread)
		Post := new(models.Post)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&Thread.Id, &Thread.Author, &Thread.Created, &Thread.Forum, &Thread.Message, &sqlSlug, &Thread.Title, &Thread.Votes)

		if !sqlSlug.Valid {
			Thread.Slug = ""
		} else {
			Thread.Slug = sqlSlug.String
		}

		postDetail.Thread = Thread
		postDetail.Post = Post

	} else if !relatedUsr && relatedThr && relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, thrs.id, thrs.author, thrs.created, thrs.forum, thrs.message, thrs.slug, thrs.title, thrs.votes,  frm.posts, frm.slug, frm.threads, frm.title, frm.author  FROM Post psts " +
			"JOIN Thread thrs ON psts.id=$1 AND psts.thread=thrs.id JOIN Forum frm ON psts.forum=frm.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		Thread := new(models.Thread)
		Forum := new(models.Forum)
		Post := new(models.Post)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&Thread.Id, &Thread.Author, &Thread.Created, &Thread.Forum, &Thread.Message, &sqlSlug, &Thread.Title, &Thread.Votes,
			&Forum.Posts, &Forum.Slug, &Forum.Threads, &Forum.Title, &Forum.User)

		if !sqlSlug.Valid {
			Thread.Slug = ""
		} else {
			Thread.Slug = sqlSlug.String
		}

		postDetail.Thread = Thread
		postDetail.Post = Post
		postDetail.Forum = Forum

	} else if relatedUsr && !relatedThr && !relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, usrs.about, usrs.email, usrs.fullname, usrs.nickname FROM Post psts " +
			"JOIN Users usrs ON psts.id=$1 AND usrs.nickname=psts.author"
		row := db.DbQueryRow(query, []interface{}{id})

		User := new(models.User)
		Post := new(models.Post)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&User.About, &User.Email, &User.Fullname, &User.Nickname)

		postDetail.Author = User
		postDetail.Post = Post

	} else if relatedUsr && !relatedThr && relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, usrs.about, usrs.email, usrs.fullname, usrs.nickname, frm.posts, frm.slug, frm.threads, frm.title, frm.author   FROM Post psts " +
			"JOIN Users usrs ON psts.id=$1 AND usrs.nickname=psts.author " +
			"JOIN Forum frm ON psts.forum=frm.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		User := new(models.User)
		Post := new(models.Post)
		Forum := new(models.Forum)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&User.About, &User.Email, &User.Fullname, &User.Nickname,
			&Forum.Posts, &Forum.Slug, &Forum.Threads, &Forum.Title, &Forum.User)

		postDetail.Author = User
		postDetail.Post = Post
		postDetail.Forum = Forum

	} else if relatedUsr && relatedThr && !relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, usrs.about, usrs.email, usrs.fullname, usrs.nickname,  thrs.id, thrs.author, thrs.created, thrs.forum, thrs.message, thrs.slug, thrs.title, thrs.votes FROM Post psts " +
			"JOIN Users usrs ON psts.id=$1 AND usrs.nickname=psts.author " +
			"JOIN Thread thrs ON psts.thread=thrs.id"
		row := db.DbQueryRow(query, []interface{}{id})

		User := new(models.User)
		Post := new(models.Post)
		Thread := new(models.Thread)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&User.About, &User.Email, &User.Fullname, &User.Nickname,
			&Thread.Id, &Thread.Author, &Thread.Created, &Thread.Forum, &Thread.Message, &sqlSlug, &Thread.Title, &Thread.Votes)

		if !sqlSlug.Valid {
			Thread.Slug = ""
		} else {
			Thread.Slug = sqlSlug.String
		}

		postDetail.Author = User
		postDetail.Post = Post
		postDetail.Thread = Thread

	} else if relatedUsr && relatedThr && relatedFrm {

		query := "SELECT psts.author,psts.created,psts.forum,psts.id,psts.isedited,psts.message,psts.parent,psts.thread, usrs.about, usrs.email, usrs.fullname, usrs.nickname,  thrs.id, thrs.author, thrs.created, thrs.forum, thrs.message, thrs.slug, thrs.title, thrs.votes , frm.posts, frm.slug, frm.threads, frm.title, frm.author FROM Post psts " +
			"JOIN Users usrs ON psts.id=$1 AND usrs.nickname=psts.author " +
			"JOIN Thread thrs ON psts.thread=thrs.id " +
			"JOIN Forum frm ON psts.forum=frm.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		User := new(models.User)
		Post := new(models.Post)
		Thread := new(models.Thread)
		Forum := new(models.Forum)

		err = row.Scan(&Post.Author, &Post.Created, &Post.Forum, &Post.Id, &Post.IsEdited, &Post.Message, &Post.Parent, &Post.Thread,
			&User.About, &User.Email, &User.Fullname, &User.Nickname,
			&Thread.Id, &Thread.Author, &Thread.Created, &Thread.Forum, &Thread.Message, &sqlSlug, &Thread.Title, &Thread.Votes,
			&Forum.Posts, &Forum.Slug, &Forum.Threads, &Forum.Title, &Forum.User)

		if !sqlSlug.Valid {
			Thread.Slug = ""
		} else {
			Thread.Slug = sqlSlug.String
		}

		postDetail.Author = User
		postDetail.Post = Post
		postDetail.Thread = Thread
		postDetail.Forum = Forum

	}

	if err != nil {
		// fmt.Println(err.Error())
		Errors.SendError("Can't find Post with id "+id, http.StatusNotFound, &w)
		return
	}

	resData, _ := postDetail.MarshalJSON()
	w.Write(resData)
	return
}
