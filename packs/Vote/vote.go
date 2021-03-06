package Vote

import (
	// "forumDb/db"
	// "forumDb/models"
	// "forumDb/packs/Errors"
	// "forumDb/packs/Thread"

	"github.com/kirillsonk/forumDb/db"
	"github.com/kirillsonk/forumDb/models"
	"github.com/kirillsonk/forumDb/packs/Errors"
	"github.com/kirillsonk/forumDb/packs/Thread"

	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func VoteThread(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		vars := mux.Vars(r)
		slugOrId := vars["slug_or_id"]
		w.Header().Set("content-type", "application/json")

		vote := models.Vote{}

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dbConn := db.GetLink()

		dbc, err := dbConn.Begin()

		if err != nil {
			// fmt.Println("db.begin ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer dbc.Rollback()

		_, err = dbc.Exec("SET LOCAL synchronous_commit TO OFF")

		if err != nil {
			// fmt.Println("set local ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = vote.UnmarshalJSON(body)
		threadId, err := strconv.Atoi(slugOrId)

		if err != nil {
			_, err = dbc.Exec("INSERT INTO Vote(nickname, voice, thread) VALUES ($1,$2, (SELECT id FROM Thread WHERE slug=$3)) "+
				"ON CONFLICT (nickname, thread) DO "+
				"UPDATE SET voice=$2",
				vote.Nickname, vote.Voice, slugOrId)
		} else {
			_, err = dbc.Exec("INSERT INTO Vote(nickname, voice, thread) VALUES ($1,$2,$3) "+
				"ON CONFLICT (nickname, thread) DO "+
				"UPDATE SET voice=$2",
				vote.Nickname, vote.Voice, threadId)
		}

		if err != nil {
			dbc.Rollback()
			Errors.SendError("Can't find user with id "+slugOrId, http.StatusNotFound, &w)
			return

		}

		dbc.Commit()

		thread, err := Thread.GetThreadByIdOrSlug(slugOrId)

		if err != nil {
			Errors.SendError("Can't find thread with id "+slugOrId, http.StatusNotFound, &w)
			return
		}

		resData, _ := thread.MarshalJSON()
		w.Write(resData)
		return
	}
	return
}
