package Post

import (
	"ForumsApi/db"
	"ForumsApi/internal/Errors"
	"ForumsApi/internal/Thread"
	"ForumsApi/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"strings"
)

func PostCreate(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//globalCount++
	//fmt.Println(globalCount)

	//if globalCount == 15500 {
	//	db.Exec("VACUUM ANALYZE;")
	//}

	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	posts := make([]models.Post, 0)

	err = json.Unmarshal(body, &posts)


	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbConn := db.GetLink()
	t, err := dbConn.Begin()

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

	thr, err := Thread.GetThread(slugOrId)

	if err != nil{
		Errors.SendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
		return
	}

	defer t.Rollback()
	if len(posts) == 0 {
		data := make([]models.Post,0)

		resp, err := json.Marshal(data)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json")

		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
		return
	}



	//var firstCreated time.Time
	//var err error
	//stmt, err := t.Prepare("INSERT INTO posts(author, forum, message, parent, thread, created) VALUES ($1,$2,$3,$4,$5,$6) RETURNING author,created,forum,id,isedited,message,parent,thread")


	resQuery :=  "INSERT INTO posts(author, forum, message, parent, thread) VALUES "

	forumUsersInsert := "INSERT INTO forum_users(forum,author) VALUES "

	var subQuery []string
	var forumUsersSubQuery []string

	for _, p := range posts{


		values := fmt.Sprintf("('%s', '%s', '%s', %d, %d) ", p.Author, thr.Forum, p.Message, p.Parent, thr.Id)

		//query += subQuery

		subQuery = append(subQuery, values)

		//newPost := models.Post{}
		//if count == 0 { // Для того, чтобы все последующие добавления постов происхдили с той же датой и временем.
		//	row := t.QueryRow("INSERT INTO posts(author, forum, message, parent, thread) VALUES ($1,$2,$3,$4,$5) RETURNING author,created,forum,id,isedited,message,parent,thread",
		//		p.Author, thr.Forum,p.Message, p.Parent, thr.Id)
		//	err = row.Scan(&newPost.Author, &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
		//		&newPost.Parent, &newPost.Thread)
		//
		//	firstCreated = newPost.Created
		//} else {
		//	row := stmt.QueryRow(p.Author, thr.Forum,p.Message, p.Parent, thr.Id, firstCreated)
		//	err = row.Scan(&newPost.Author,  &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
		//		&newPost.Parent, &newPost.Thread)
		//}

		//_,err := t.Exec("INSERT INTO forum_users(forum,author) VALUES ($1,$2) ON CONFLICT DO NOTHING", thr.Forum, p.Author)


		forumUsersValues := fmt.Sprintf("('%s', '%s') ", thr.Forum, p.Author)

		forumUsersSubQuery = append(forumUsersSubQuery, forumUsersValues)

		//if err != nil {
		//	fmt.Println("postCreate insert forum_users ", err.Error())
		//}

		//data = append(data, newPost)


	}

	resQuery += strings.Join(subQuery, ",") + " RETURNING author,created,forum,id,isedited,message,parent,thread;"

	forumUsersInsert += strings.Join(forumUsersSubQuery, ",") + " ON CONFLICT DO NOTHING;"

	resQuery += forumUsersInsert

	//fmt.Println(resQuery)

	rows, err := t.Query(resQuery)

	if err != nil {
		//fmt.Println(err.Error())
		errorName := err.(*pq.Error).Code.Name()

		fmt.Println(errorName)
		if err.Error() == "pq: Parent post exc" {
			Errors.SendError("Parent post was created in another thread \n", 409, &w)
			return
		}

		if errorName == "foreign_key_violation" {
			Errors.SendError("Can't find parent post \n", 404, &w)
			return
		}

		if errorName != "syntax_error" {
			Errors.SendError("Can't find parent post \n", 404, &w)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//_, err = db.Exec(forumUsersInsert)
	//
	//if err != nil {
	//	fmt.Println("forumUsersInsert ---- ", err.Error())
	//	return
	//}



	data := make([]models.Post,0)

	for rows.Next() {
		newPost := models.Post{}

		err := rows.Scan(&newPost.Author,  &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
			&newPost.Parent, &newPost.Thread)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data = append(data, newPost)
	}

	resp, err := json.Marshal(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Commit()

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)

	return
}


func PostDetails(w http.ResponseWriter, r *http.Request){

	vars := mux.Vars(r)
	id := vars["id"]

	related := r.URL.Query().Get("related")

	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		post := new(models.Post)

		err = json.Unmarshal(body, post)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err != nil{
			Errors.SendError( "Can't find post with id " + id + "\n", 404, &w)
			return
		}

		if post.Message == "" {
			row := db.DbQueryRow("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE id=$1", []interface{}{id})

			err = row.Scan(&post.Author,&post.Created,&post.Forum,&post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

			if err != nil{
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			resp, _ := json.Marshal(post)
			w.Header().Set("content-type", "application/json")

			w.Write(resp)

			return
		}
		row := db.DbQueryRow("UPDATE posts SET message=$1, isedited=true WHERE id=$2 RETURNING author,created,forum,id,isedited,message,parent,thread", []interface{}{post.Message, id})
		err = row.Scan(&post.Author,&post.Created,&post.Forum,&post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

		if err != nil {
			Errors.SendError("Can't find post with id "+id+"\n", 404, &w)
			return
		}

		resp, _ := json.Marshal(post)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}
	postDetail := models.PostDetail{}

	//query := "SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE id=$1; "


	//post := new(models.Post)
	//row := db.QueryRow("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE id=$1;", id)
	//
	//err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
	//
	//if err != nil {
	//	Errors.SendError("Can't find post with id "+id+"\n", 404, &w)
	//	return
	//}
	//
	//postDetail.Post = post

	//var query string

	var oblectsArr []string
	objects := strings.Split(related, ",")
	for index := range objects  {
		item := objects[index]
		oblectsArr = append(oblectsArr, item)
	}

	var relUser = false
	var relThread = false
	var relForum = false

	for index := range oblectsArr {
		if oblectsArr[index] == "user" {
			relUser = true
			//query += "SELECT about,email,fullname,nickname FROM users WHERE nickname='" + post.Author + "'; "
			//author, _ := getUser(post.Author, nil)
			//postDetail.Author = author
		}
		if oblectsArr[index] == "thread" {
			relThread = true
			//_, err := strconv.Atoi(strconv.Itoa(int(post.Thread)))
			//
			//if err != nil {
			//	query += "SELECT * FROM threads WHERE slug='" + strconv.Itoa(int(post.Thread)) + "'; "
			//} else {
			//	query += "SELECT * FROM threads WHERE id=" + strconv.Itoa(int(post.Thread)) + "; "
			//}

			//if err != nil {
			//	w.WriteHeader(http.StatusInternalServerError)
			//	return
			//}

			//postDetail.Thread = thread
		}
		if oblectsArr[index] == "forum" {
			relForum = true
			//query += "SELECT * FROM forums WHERE slug='" + post.Forum + "'; "
			//forum, _ := getForum(post.Forum, nil)
			//postDetail.Forum = forum
		}
	}

	var err error
	var sqlSlug sql.NullString

	if !relUser && !relThread && !relForum {
		post := new(models.Post)
		row := db.DbQueryRow("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE id=$1;", []interface{}{id})

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

		postDetail.Post = post

	} else if !relUser && !relThread && relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, f.posts, f.slug, f.threads, f.title, f.author FROM posts p " +
			"JOIN forums f ON p.id=$1 AND p.forum=f.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		forum := new(models.Forum)
		post := new(models.Post)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

		postDetail.Forum = forum
		postDetail.Post = post

	} else if !relUser && relThread && !relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, t.id, t.author, t.created, t.forum, t.message, t.slug, t.title, t.votes FROM posts p " +
			"JOIN threads t ON p.id=$1 AND p.thread=t.id"
		row := db.DbQueryRow(query, []interface{}{id})

		thr := new(models.Thread)
		post := new(models.Post)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &sqlSlug, &thr.Title, &thr.Votes)

		if !sqlSlug.Valid {
			thr.Slug = ""
		} else {
			thr.Slug = sqlSlug.String
		}


		postDetail.Thread = thr
		postDetail.Post = post

	} else if !relUser && relThread && relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, t.id, t.author, t.created, t.forum, t.message, t.slug, t.title, t.votes,  f.posts, f.slug, f.threads, f.title, f.author  FROM posts p " +
			"JOIN threads t ON p.id=$1 AND p.thread=t.id JOIN forums f ON p.forum=f.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		thr := new(models.Thread)
		forum := new(models.Forum)
		post := new(models.Post)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &sqlSlug, &thr.Title, &thr.Votes,
			&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

		if !sqlSlug.Valid {
			thr.Slug = ""
		} else {
			thr.Slug = sqlSlug.String
		}


		postDetail.Thread = thr
		postDetail.Post = post
		postDetail.Forum = forum

	} else if relUser && !relThread && !relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, u.about, u.email, u.fullname, u.nickname FROM posts p " +
			"JOIN users u ON p.id=$1 AND u.nickname=p.author"
		row := db.DbQueryRow(query, []interface{}{id})

		user := new(models.User)
		post := new(models.Post)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&user.About, &user.Email, &user.FullName, &user.NickName)

		postDetail.Author = user
		postDetail.Post = post

	} else if relUser && !relThread && relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, u.about, u.email, u.fullname, u.nickname, f.posts, f.slug, f.threads, f.title, f.author   FROM posts p " +
			"JOIN users u ON p.id=$1 AND u.nickname=p.author " +
			"JOIN forums f ON p.forum=f.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		user := new(models.User)
		post := new(models.Post)
		forum := new(models.Forum)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&user.About, &user.Email, &user.FullName, &user.NickName,
			&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

		postDetail.Author = user
		postDetail.Post = post
		postDetail.Forum = forum

	} else if relUser && relThread && !relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, u.about, u.email, u.fullname, u.nickname,  t.id, t.author, t.created, t.forum, t.message, t.slug, t.title, t.votes FROM posts p " +
			"JOIN users u ON p.id=$1 AND u.nickname=p.author " +
			"JOIN threads t ON p.thread=t.id"
		row := db.DbQueryRow(query, []interface{}{id})

		user := new(models.User)
		post := new(models.Post)
		thr := new(models.Thread)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&user.About, &user.Email, &user.FullName, &user.NickName,
			&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &sqlSlug, &thr.Title, &thr.Votes)

		if !sqlSlug.Valid {
			thr.Slug = ""
		} else {
			thr.Slug = sqlSlug.String
		}


		postDetail.Author = user
		postDetail.Post = post
		postDetail.Thread = thr

	} else if relUser && relThread && relForum {

		query := "SELECT p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread, u.about, u.email, u.fullname, u.nickname,  t.id, t.author, t.created, t.forum, t.message, t.slug, t.title, t.votes , f.posts, f.slug, f.threads, f.title, f.author FROM posts p " +
			"JOIN users u ON p.id=$1 AND u.nickname=p.author " +
			"JOIN threads t ON p.thread=t.id " +
			"JOIN forums f ON p.forum=f.slug"
		row := db.DbQueryRow(query, []interface{}{id})

		user := new(models.User)
		post := new(models.Post)
		thr := new(models.Thread)
		forum := new(models.Forum)

		err = row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread,
			&user.About, &user.Email, &user.FullName, &user.NickName,
			&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &sqlSlug, &thr.Title, &thr.Votes,
			&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

		if !sqlSlug.Valid {
			thr.Slug = ""
		} else {
			thr.Slug = sqlSlug.String
		}


		postDetail.Author = user
		postDetail.Post = post
		postDetail.Thread = thr
		postDetail.Forum = forum

	}


	if err != nil {
		fmt.Println(err.Error())
		Errors.SendError("Can't find post with id "+id+"\n", 404, &w)
		return
	}

	//rows, err := db.Query(query)

	//fmt.Println(query)

	//if err != nil {
	//	Errors.SendError("Can't find post with id "+id+"\n", 404, &w)
	//	return
	//}


	//fmt.Println("before")
	//for rows.Next(){
	//	fmt.Println("after")
	//
	//	if relUser {
	//		fmt.Println("get user")
	//
	//		user := new(models.User)
	//		err := rows.Scan(&user.About, &user.Email, &user.FullName, &user.NickName)
	//		if err != nil {
	//			fmt.Println(err.Error())
	//			Errors.SendError("user", 500, &w)
	//			return
	//		}
	//
	//		postDetail.Author = user
	//
	//		relUser = false
	//	} else if relThread {
	//		fmt.Println("get thread")
	//
	//		thr := new(models.Thread)
	//		err = rows.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &thr.Slug, &thr.Title, &thr.Votes)
	//
	//		if err != nil {
	//			Errors.SendError("thread", 500, &w)
	//			return
	//		}
	//
	//		postDetail.Thread = thr
	//
	//		relThread = false
	//	} else if relForum {
	//		fmt.Println("get forum")
	//
	//		forum := new(models.Forum)
	//		err = rows.Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
	//
	//		if err != nil {
	//			Errors.SendError("forum", 500, &w)
	//			return
	//		}
	//
	//		postDetail.Forum = forum
	//
	//		relForum = false
	//	}
	//
	//}




	//row := db.QueryRow("SELECT author,created,forum,id,isedited,message,parent,thread FROM posts WHERE id=$1", id)

	//post := models.Post{}

	//err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

	//if err != nil {
	//	Errors.SendError( "Can't find post with id " + id + "\n", 404, &w)
	//	return
	//}



	resp, _ := json.Marshal(postDetail)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)
	return
}

