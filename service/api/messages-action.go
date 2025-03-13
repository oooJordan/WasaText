package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) sendNewMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// recupero il conversation_id dai parametri della richiesta
	conversationIDStr := ps.ByName("conversation_id")
	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// decodifico il corpo della richiesta per ottenere il messaggio
	var message MessageSent
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// controllo se il formato del messaggio è corretto
	if message.Media != "text" && message.Media != "gif" && message.Media != "gif_with_text" {
		http.Error(w, "Invalid message type", http.StatusBadRequest)
		return
	}

	// recupero il tipo di conversazione dal database
	conversationType, err := rt.db.GetConversationType(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get conversation type")
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// controllo se l'utente è nella conversazione
	if conversationType == "private_chat" {
		isParticipant, err := rt.db.IsUserInPrivateChat(conversationID, author)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to check private chat membership")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !isParticipant {
			http.Error(w, "User is not a participant of this private chat", http.StatusNotFound)
			return
		}
	} else if conversationType == "group_chat" {
		isParticipant, err := rt.db.IsUserInGroup(conversationID, author)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to check private chat membership")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !isParticipant {
			http.Error(w, "User is not a participant of this private chat", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "Invalid conversation type", http.StatusBadRequest)
		return
	}

	// aggiungo il messaggio nel database
	messageID, err := rt.db.NewMessage(conversationID, author, message.Media, message.Content, message.Image)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to create message")
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// rispondo con il nuovo messageId generato
	response := map[string]int{
		"messageId": messageID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Verifico se l'utente è autorizzato
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		// Se l'utente non è autorizzato, restituisco un errore 401
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Recupero i parametri della richiesta (conversation_id e message_id)
	destConvIDStr := ps.ByName("conversation_id")
	origMsgIDStr := ps.ByName("message_id")

	// Converto i parametri in interi per poterli utilizzare nelle query
	destinationConversationID, err := strconv.Atoi(destConvIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	originalMessageID, err := strconv.Atoi(origMsgIDStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}
	// Inoltro il messaggio e recupero l'ID del messaggio inoltrato dal database
	forwardedMessageID, err := rt.db.ForwardMessage(destinationConversationID, originalMessageID, author)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Original message or destination conversation not found", http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("Failed to forward message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Se il messaggio è stato inoltrato con successo, restituisco l'ID del messaggio inoltrato
	response := map[string]interface{}{
		"forwardedMessageId": forwardedMessageID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
}
