package main

import (
	"fmt"
	"forum-database/db"
	"forum-database/internal/CommonService"
	"forum-database/internal/Forum"
	"forum-database/internal/Post"
	"forum-database/internal/Thread"
	"forum-database/internal/User"
	"forum-database/internal/Vote"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func AccessLogMiddleware(mux *mux.Router) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mux.ServeHTTP(w, r)

		sortVal := r.URL.Query().Get("sort")
		if sortVal != "" {
			fmt.Println("method ", r.Method, "; url", r.URL.Path, " Sort: ", sortVal,
				"Time work: ", time.Since(start))
		} else {
			fmt.Println("method ", r.Method, "; url", r.URL.Path,
				"Time work: ", time.Since(start))

		}
	})
}

func main() {
	postgres, _ := db.InitDb()

	router := mux.NewRouter()

	router.HandleFunc("/api/forum/create", Forum.CreateForum)
	router.HandleFunc(`/api/forum/{slug}/create`, Thread.ThreadCreate)
	router.HandleFunc(`/api/forum/{slug}/details`, Forum.ForumDetails)
	router.HandleFunc(`/api/forum/{slug}/threads`, Forum.ThreadsForum)
	router.HandleFunc(`/api/forum/{slug}/users`, Forum.UsersForum)

	router.HandleFunc(`/api/post/{id}/details`, Post.PostDetails)

	router.HandleFunc(`/api/service/clear`, CommonService.ServiceClear)
	router.HandleFunc(`/api/service/status`, CommonService.ServiceStatus)

	router.HandleFunc(`/api/thread/{slug_or_id}/create`, Post.CreatePost)
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, Thread.ThreadDetails)
	router.HandleFunc(`/api/thread/{slug_or_id}/posts`, Thread.ThreadPosts)
	router.HandleFunc(`/api/thread/{slug_or_id}/vote`, Vote.ThreadVote)

	router.HandleFunc(`/api/user/{nickname}/create`, User.UserCreate)
	router.HandleFunc(`/api/user/{nickname}/profile`, User.UserProfile)

	//	siteHandler := AccessLogMiddleware(router)

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)

	defer postgres.Close()

	return
}
