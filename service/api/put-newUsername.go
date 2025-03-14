package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
	"github.com/oooJordan/WasaText/service/database"
)

// #AGGIORNO USERNAME UTENTE#
func (rt *_router) UpdateUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Controllo se il token è valido
	isValid, author, err := rt.IsValidToken(r, w)
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

	// Aggiorno l'username nel database
	err = rt.db.UpdateUsername(author, newUsername.NewUsername)
	if err != nil {
		ctx.Logger.Error(err)
		// Se l'errore è dovuto al conflitto di username, restituisco errore 409
		if errors.Is(err, database.ErrUsernameAlreadyInUse) {
			http.Error(w, "Username already in use", http.StatusConflict) // 409
			return
		}
		// 500
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	// Successo
	w.WriteHeader(http.StatusNoContent)
}
