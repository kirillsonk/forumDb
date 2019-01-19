package main

import (
	// "forumDb/db"
	// "forumDb/packs/CommonService"
	// "forumDb/packs/Forum"
	// "forumDb/packs/Post"
	// "forumDb/packs/Thread"
	// "forumDb/packs/User"
	// "forumDb/packs/Vote"

	"net/http"

	"github.com/kirillsonk/forumDb/db"
	"github.com/kirillsonk/forumDb/packs/CommonService"
	"github.com/kirillsonk/forumDb/packs/Forum"
	"github.com/kirillsonk/forumDb/packs/Post"
	"github.com/kirillsonk/forumDb/packs/Thread"
	"github.com/kirillsonk/forumDb/packs/User"
	"github.com/kirillsonk/forumDb/packs/Vote"

	"github.com/gorilla/mux"
)

func main() {
	postgres, _ := db.InitDatabase()
	
	dbConnection, _ := db.InitDbSQL()

	router := mux.NewRouter()

	router.HandleFunc(`/api/user/{nickname}/create`, User.CreateUser)
	router.HandleFunc(`/api/user/{nickname}/profile`, User.UserProfile)
	router.HandleFunc(`/api/service/clear`, CommonService.ClearService)
	router.HandleFunc(`/api/service/status`, CommonService.ServiceStatus)
	router.HandleFunc("/api/forum/create", Forum.CreateForum)
	router.HandleFunc(`/api/forum/{slug}/create`, Thread.ThreadCreate)
	router.HandleFunc(`/api/forum/{slug}/details`, Forum.ForumDetails)
	router.HandleFunc(`/api/forum/{slug}/threads`, Forum.ThreadsForum)
	router.HandleFunc(`/api/forum/{slug}/users`, Forum.UsersForum)
	router.HandleFunc(`/api/post/{id}/details`, Post.PostDetails)
	router.HandleFunc(`/api/thread/{slug_or_id}/create`, Post.CreatePost)
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, Thread.ThreadDetails)
	router.HandleFunc(`/api/thread/{slug_or_id}/posts`, Thread.PostsThread)
	router.HandleFunc(`/api/thread/{slug_or_id}/vote`, Vote.VoteThread)

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)

	defer postgres.Close()
	defer dbConnection.Close()
	return
}
