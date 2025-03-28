package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// ------------------------------ #LASCIA IL GRUPPO# ------------------------
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
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
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

// ------------------------- #AGGIORNO L'USERNAME DEL GRUPPO# -----------------------
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
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
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

	// aggiorno il nome del gruppo
	err = rt.db.UpdateGroupName(conversationID, newName.NewUsername)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to update group name")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Risposta di successo
	w.WriteHeader(http.StatusNoContent)
}

// ------------------------- #AGGIORNO IMMAGINE DEL GRUPPO# ---------------------------
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
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
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
	w.WriteHeader(http.StatusOK)
}

// ---------------------- #AGGIUNGO MEMBRI A UN GRUPPO# ----------------------------------
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
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
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
