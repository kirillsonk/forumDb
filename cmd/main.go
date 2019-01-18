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

	"github.com/kirillsonk/forumDb/db"
	"github.com/kirillsonk/forumDb/internal/CommonService"
	"github.com/kirillsonk/forumDb/internal/Forum"
	"github.com/kirillsonk/forumDb/internal/Post"
	"github.com/kirillsonk/forumDb/internal/Thread"
	"github.com/kirillsonk/forumDb/internal/User"
	"github.com/kirillsonk/forumDb/internal/Vote"

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
