package main

import (
	"ForumsApi/db"

	"ForumsApi/internal/CommonService"
	"ForumsApi/internal/Forum"
	"ForumsApi/internal/Post"
	"ForumsApi/internal/Thread"
	"ForumsApi/internal/User"
	"ForumsApi/internal/Vote"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)


func AccessLogMiddleware (mux *mux.Router,) http.HandlerFunc   {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()

		mux.ServeHTTP(w, r)

		sortVal := r.URL.Query().Get("sort")
		if sortVal != "" {
			fmt.Println("method ", r.Method, "; url", r.URL.Path,  " Sort: ", sortVal,
				"Time work: ", time.Since(begin))
		} else {
			fmt.Println("method ", r.Method, "; url", r.URL.Path,
				"Time work: ", time.Since(begin))

		}

		//if sortVal != "" {
		//	fmt.Println("END method ", r.Method, " Sort: ", sortVal, "; url", r.URL.Path,
		//		"Time work: ", time.Since(begin))
		//} else {
		//	fmt.Println("END method ", r.Method, "; url", r.URL.Path,
		//		"Time work: ", time.Since(begin))
		//}

	})
}

func main(){
	postgres, _ := db.InitDb()

	router := mux.NewRouter()

	router.HandleFunc("/api/forum/create", Forum.ForumCreate)
	router.HandleFunc(`/api/forum/{slug}/create`, Thread.ThreadCreate)
	router.HandleFunc(`/api/forum/{slug}/details`, Forum.ForumDetails) // +
	router.HandleFunc(`/api/forum/{slug}/threads`, Forum.ForumThreads) // - не оч
	router.HandleFunc(`/api/forum/{slug}/users`, Forum.ForumUsers) // +

	router.HandleFunc(`/api/post/{id}/details`, Post.PostDetails) // +

	router.HandleFunc(`/api/service/clear`, CommonService.ServiceClear)
	router.HandleFunc(`/api/service/status`, CommonService.ServiceStatus) // -

	router.HandleFunc(`/api/thread/{slug_or_id}/create`, Post.PostCreate)
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, Thread.ThreadDetails) // +
	router.HandleFunc(`/api/thread/{slug_or_id}/posts`, Thread.ThreadPosts) // +
	router.HandleFunc(`/api/thread/{slug_or_id}/vote`, Vote.ThreadVote)

	router.HandleFunc(`/api/user/{nickname}/create`, User.UserCreate)
	router.HandleFunc(`/api/user/{nickname}/profile`, User.UserProfile)  // + быстро

//	siteHandler := AccessLogMiddleware(router)

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)

	defer postgres.Close()

	return
}

