package Vote

import (
	"forum-database/db"
	"forum-database/internal/Errors"
	"forum-database/internal/Thread"
	"forum-database/models"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func ThreadVote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if r.Method != http.MethodPost {
		return
	}

	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	vote := models.Vote{}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

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

	defer t.Rollback()

	_, err = t.Exec("SET LOCAL synchronous_commit TO OFF")

	if err != nil {
		fmt.Println("set local ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = vote.UnmarshalJSON(body)
	thrId, err := strconv.Atoi(slugOrId)

	if err != nil {
		_, err = t.Exec("INSERT INTO votes(nickname, voice, thread) VALUES ($1,$2, (SELECT id FROM threads WHERE slug=$3)) "+
			"ON CONFLICT (nickname, thread) DO "+
			"UPDATE SET voice=$2",
			vote.Nickname, vote.Voice, slugOrId)
	} else {
		_, err = t.Exec("INSERT INTO votes(nickname, voice, thread) VALUES ($1,$2,$3) "+
			"ON CONFLICT (nickname, thread) DO "+
			"UPDATE SET voice=$2",
			vote.Nickname, vote.Voice, thrId)
	}

	if err != nil {
		if err.(*pq.Error).Code.Name() == "foreign_key_violation" {
			Errors.SendError("Can't find user with id "+slugOrId+"\n", 404, &w)
			return
		}
	}

	t.Commit()

	thr, err := Thread.GetThread(slugOrId)

	if err != nil {
		Errors.SendError("Can't find thread with id "+slugOrId+"\n", 404, &w)
		return
	}

	resData, _ := thr.MarshalJSON()
	w.Write(resData)
	return

}
