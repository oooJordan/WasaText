package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) updateProfileImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	var NewImage UpdateProfileImageRequest
	err = json.NewDecoder(r.Body).Decode(&NewImage)
	if err != nil {
		ctx.Logger.WithError(err).Error("error decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Controllo validità dell'immagine (es. URL non vuoto)
	if NewImage.Image == "" {
		http.Error(w, "Image URL cannot be empty", http.StatusBadRequest)
		return
	}
	// Aggiorna l'immagine di profilo nel database
	err = rt.db.UpdateProfileImage(author, NewImage.Image)
	if err != nil {
		ctx.Logger.Error(err)
		// Se si verifica un errore generico, restituisci 500
		http.Error(w, "Failed to update profile image", http.StatusInternalServerError)
		return
	}

	// Successo
	w.WriteHeader(http.StatusNoContent)
}
