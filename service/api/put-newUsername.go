package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) UpdateUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userIDStr := ps.ByName("user_id")
	userIDint, err := strconv.Atoi(userIDStr)
	if err != nil {
		ctx.Logger.WithError(err).Error("error converting uid to int")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Controllo se il token è valido
	isValid, err := rt.IsValidToken(r, w)
	if err != nil {
		// La risposta HTTP è già gestita all'interno di IsValidToken
		return
	}
	if !isValid {
		// Token non valido
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newUsername UpdateUsernameRequest
	err = json.NewDecoder(r.Body).Decode(&newUsername)
	if err != nil {
		ctx.Logger.WithError(err).Error("error decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidNickname(newUsername.NewUsername) {
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	// Aggiorna l'username nel database
	if err := rt.db.UpdateUsername(userIDint, newUsername.NewUsername); err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	// Successo
	w.WriteHeader(http.StatusNoContent)
}
