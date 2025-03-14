package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// #OTTENGO LA LISTA DEGLI UTENTI#
func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	username := r.URL.Query().Get("name")

	isValid, _, err := rt.IsValidToken(r, w)
	if err != nil {
		// la risposta HTTP è già gestita all'interno di IsValidToken
		return
	}
	if !isValid {
		// token non valido
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := rt.db.GetListUsers(username)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error executing query")
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	// controllo che sia stato trovato almeno un utente
	if len(users) == 0 {
		w.WriteHeader(http.StatusNotFound) // 404
		return
	}

	// codifico la lista degli utenti in JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		ctx.Logger.WithError(err).Error("Error encoding response")
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}
