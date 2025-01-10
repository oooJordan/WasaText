package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) renameGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// controllo se il token è valido
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
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

	var newName UpdateUsernameRequest
	err = json.NewDecoder(r.Body).Decode(&newName)
	if err != nil || newName.NewUsername == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

	// Aggiorna il nome del gruppo
	err = rt.db.UpdateGroupName(conversationID, newName.NewUsername)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to update group name")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Risposta di successo
	w.WriteHeader(http.StatusNoContent)
}
