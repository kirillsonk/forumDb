package CommonService

import (
	// "forum-database/db"
	// "forum-database/models"
	"net/http"

	"bitbucket.org/kirillsonk/forum-database/db"
	"bitbucket.org/kirillsonk/forum-database/models"
)

func ServiceStatus(w http.ResponseWriter, r *http.Request) { //правил if для метода GET
	if r.Method == http.MethodGet {
		w.Header().Set("content-type", "application/json")

		row := db.DbQueryRow("SELECT t1.cnt c1, t2.cnt c2, t3.cnt c3, t4.cnt c4 FROM (SELECT count(*) cnt FROM Users) t1, (SELECT COUNT(*) cnt FROM Forum) t2, (SELECT COUNT(*) cnt FROM Post) t3, (SELECT COUNT(*) cnt FROM Thread) t4", nil)
		status := models.Status{}
		err := row.Scan(&status.User, &status.Forum, &status.Post, &status.Thread)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resData, _ := status.MarshalJSON()
		w.WriteHeader(http.StatusOK)
		w.Write(resData)
		return
	}
	return
}

func ClearService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	db.DbExec("TRUNCATE TABLE Vote, Users, Post, Thread, Forum, ForumUser", nil)
	w.WriteHeader(http.StatusOK)
	return
}
