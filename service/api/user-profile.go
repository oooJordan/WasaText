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

	// controllo che l'url sia valido (URL non vuoto)
	if NewImage.Image == "" || !validateLocalImageURL(NewImage.Image) {
		http.Error(w, "Image URL cannot be empty and must be valid and point to an image", http.StatusBadRequest)
		return
	}

	// aggiorno l'immagine di profilo nel database
	err = rt.db.UpdateProfileImage(author, NewImage.Image)
	if err != nil {
		ctx.Logger.Error(err)
		http.Error(w, "Failed to update profile image", http.StatusInternalServerError)
		return
	}

	// Successo
	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) getProfileImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// controllo se il token è valido
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// recupero l'immagine dal database
	imageURL, err := rt.db.GetProfileImage(author)
	if err != nil {
		ctx.Logger.Error(err)
		http.Error(w, "Failed to fetch profile image", http.StatusInternalServerError)
		return
	}

	// rispondo con l'immagine
	response := map[string]string{
		"actualImage": imageURL,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
