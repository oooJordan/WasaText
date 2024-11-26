package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) updateUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	token, err := IsValidToken(r, w)
	if err != nil {
		ctx.Logger.WithError(err).Error("token: Authentication Error")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userid, _ := strconv.Atoi(token)
	var body struct {
		NewUsername string `json:"newUsername"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		ctx.Logger.WithError(err).Error("Error decoding JSON body")
		w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
		return
	}
	//controllo se vuoto
	if body.NewUsername == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// aggiorno nome utente nel database
	err = rt.db.UpdateUsername(userid, body.NewUsername)
	if err != nil {
		if errors.Is(err, errors.New("username already in use")) {
			w.WriteHeader(http.StatusConflict) // 409
			return
		}
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
