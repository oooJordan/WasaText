package api

import (
	//"encoding/json"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	username := r.URL.Query().Get("name")
	ctx.Logger.Infof("Username received: '%s'", username)
	/*
		token, err := IsValidToken(r,w)
		if err != nil{
			ctx.Logger.WithError(err).Error("token: Authentication Error")
			w.WriteHeader(http.StatusForbidden) //403
			return
		}
		userid, _ := strconv.Atoi(token)
	*/
	users, err := rt.db.GetListUsers(username)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error executing query")
		w.WriteHeader(http.StatusInternalServerError) //500
		return
	}

	if len(users) == 0 {
		w.WriteHeader(http.StatusNotFound) //404
		return
	}

	// Codifica la lista degli utenti in JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		ctx.Logger.WithError(err).Error("Error encoding response")
		w.WriteHeader(http.StatusInternalServerError) //500
	}
}
