package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) newConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	var req ConversationRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ChatType == "" || req.StartMessage.Type == "" || len(req.Usersname) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	dbReq := convertToDatabaseConversationRequest(req)

	conversationID, err := rt.db.CreateConversationDB(author, dbReq)
	if err != nil {
		ctx.Logger.Error(err)
		http.Error(w, "Error creating conversation", http.StatusInternalServerError)
		return
	}
	lastMessage := req.StartMessage

	response := CreateConversationResponse{
		ConversationID: conversationID,
		Message:        "Conversation created successfully",
		LastMessage:    lastMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
