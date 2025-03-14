package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// #AGGIUNGI MEMBRI A UN GRUPPO#
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// controllo se il token è valido
	isValid, _, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// recupero il conversation_id dai parametri
	conversationIDStr := ps.ByName("conversation_id")
	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	// recupero il nome dell'utente da aggiungere
	var newUser LoginRequest
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil || newUser.User == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// verifico che la conversazione esista e sia un gruppo
	groupExists, err := rt.db.IsGroupConversation(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to query conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !groupExists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// recupero l'ID dell'utente da aggiungere
	userID, err := rt.db.GetUserIDByUsername(newUser.User)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to query user by name")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// aggiungo l'utente al gruppo
	err = rt.db.AddUserToGroup(conversationID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to add user to group")
		http.Error(w, "Failed to add user to group", http.StatusInternalServerError)
		return
	}

	// successo
	w.WriteHeader(http.StatusNoContent)
}
