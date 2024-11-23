package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	nickname := r.URL.Query().Get("user_id") // get nickename utente

	// !!! devo controllare che abbia fatto l'accesso però mi serve il token !!!

	//verifico se il parametro è valido (Error 400)
	if nickname == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//lo cerco nel DB
	nome, err := rt.db.SearchUser(nickname)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error during searching for nickname")
		w.WriteHeader(http.StatusInternalServerError) //Error 500
		return
	}

	//utente non trovato (Error 404)
	if len(nome) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//successo: lista di utenti con quel nickname (200 OK)
	var users []User
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}
