package main

import (
	// "forumDb/db"
	// "forumDb/internal/CommonService"
	// "forumDb/internal/Forum"
	// "forumDb/internal/Post"
	// "forumDb/internal/Thread"
	// "forumDb/internal/User"
	// "forumDb/internal/Vote"

	"net/http"

	"bitbucket.org/kirillsonk/forumDb/CommonService"
	"bitbucket.org/kirillsonk/forumDb/Forum"
	"bitbucket.org/kirillsonk/forumDb/Post"
	"bitbucket.org/kirillsonk/forumDb/Thread"
	"bitbucket.org/kirillsonk/forumDb/User"
	"bitbucket.org/kirillsonk/forumDb/Vote"
	"bitbucket.org/kirillsonk/forumDb/db"

	"github.com/gorilla/mux"
)

func main() {
	postgres, _ := db.InitDatabase()
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

	return
}
