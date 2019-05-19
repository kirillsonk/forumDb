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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)


func AccessLogMiddleware (mux *mux.Router,) http.HandlerFunc   {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		mux.ServeHTTP(w, r)

		HitStat.With(prometheus.Labels{
			"url":    r.URL.Path,
			"method": r.Method,
			"code":   w.Header().Get("Status-Code"),
		}).Inc()

		rps.Add(1)
	})
}

var (
	HitStat = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ForumsApi",
			Subsystem: "hit_stat",
			Name:      "HitStat",
			Help:      "Hit info.",
		},
		[]string{
			"url",
			"method",
			"code",
		},
	)

	rps =prometheus.NewCounter(
		prometheus.CounterOpts{
		  Name: "rps",
		})
	)
	

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(HitStat)
	prometheus.MustRegister(rps)

}


func main() {
	postgres, _ := db.InitDatabase()
	
	dbConnection, _ := db.InitDbSQL()

	router := mux.NewRouter()

	http.Handle("/metrics", promhttp.Handler())

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

	siteHandler := AccessLogMiddleware(router)

	http.Handle("/", siteHandler)
	http.ListenAndServe(":5000", nil)


	defer postgres.Close()
	defer dbConnection.Close()
	return
}
