package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// verifico che la conversazione esista e che sia un gruppo
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

	// rimozione dell'utente dal gruppo
	err = rt.db.RemoveUserFromGroup(conversationID, author)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User is not a member of this group", http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("Failed to remove user from group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// controllo se il gruppo è vuoto
	isEmpty, err := rt.db.IsGroupEmpty(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to check if group is empty")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// se il gruppo è vuoto lo elimino
	if isEmpty {
		err = rt.db.DeleteGroup(conversationID)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to delete empty group")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// successo
	w.WriteHeader(http.StatusNoContent)
}
