package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// #AGGIORNA IMMAGINE DEL GRUPPO#
func (rt *_router) updateGroupImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// recupero l'id della conversazione dai parametri
	conversationIDStr := ps.ByName("conversation_id")
	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	var NewImage UpdateProfileImageRequest
	err = json.NewDecoder(r.Body).Decode(&NewImage)
	if err != nil {
		ctx.Logger.WithError(err).Error("error decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// verifico che la conversazione esista ed è un gruppo
	isGroup, err := rt.db.IsGroupConversation(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to check conversation type")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isGroup {
		http.Error(w, "Conversation is not a group", http.StatusForbidden)
		return
	}

	// verifico che l'utente sia un membro del gruppo
	isMember, err := rt.db.IsUserInGroup(conversationID, author)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to check group membership")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not a member of the group", http.StatusNotFound)
		return
	}
	// aggiorno l'immagine del gruppo
	err = rt.db.UpdateGroupImage(conversationID, NewImage.Image)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to update group image")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// successo
	w.WriteHeader(http.StatusNoContent)
}
