package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// ------------------------------ #INVIO DI UN NUOVO MESSAGGIO# --------------------------------
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
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
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
	isMember, err := rt.isUserInConversation(conversationID, author, conversationType)
	if err != nil {
		if err.Error() == "Invalid conversation type" {
			http.Error(w, "Invalid conversation type", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not a participant of this conversation", http.StatusNotFound)
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

// -------------------------------- #INOLTRO DEL MESSAGGIO# -------------------------------
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
	// Converto i parametri in interi per poterli utilizzare nelle query
	destinationConversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
		return
	}

	originalMessageID, ok := getIntParam(w, ps, "message_id")
	if !ok {
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

// ------------------------------- #ELIMINARE IL MESSAGGIO# --------------------------
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1) Controllo che l'utente sia loggato
	isValid, userID, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Recupero gli ID della conversazione e del messaggio
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
		return
	}

	messageID, ok := getIntParam(w, ps, "message_id")
	if !ok {
		return
	}

	// 3) Verifico se la conversazione esiste
	chatType, err := rt.db.GetConversationType(conversationID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if chatType == "" {
		http.Error(w, "Conversation Not Found", http.StatusNotFound)
		return
	}

	// 4) Verifico se l'utente fa parte della conversazione
	isMember, err := rt.isUserInConversation(conversationID, userID, chatType)
	if err != nil {
		if err.Error() == "Invalid conversation type" {
			http.Error(w, "Invalid conversation type", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not a participant of this conversation", http.StatusNotFound)
		return
	}

	// 5) Controllo che il messaggio esista e sia stato inviato dall'utente
	senderID, err := rt.db.GetMessageSender(messageID, conversationID)
	if err != nil {
		if err.Error() == "message not found" {
			http.Error(w, "Message Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if senderID != userID {
		http.Error(w, "Forbidden: You can only delete your own messages", http.StatusForbidden)
		return
	}

	// 6) Elimino le reazioni al messaggio dalla tabella message_reactions
	err = rt.db.RemoveAllReaction(messageID)
	if err != nil {
		http.Error(w, "Failed to delete message reactions", http.StatusInternalServerError)
		return
	}
	// 7) Elimino prima lo stato del messaggio nella tabella messages_read_status
	err = rt.db.DeleteMessageStatus(messageID)
	if err != nil {
		http.Error(w, "Failed to delete message status", http.StatusInternalServerError)
		return
	}

	// 8) Elimino il messaggio dalla tabella messages
	err = rt.db.DeleteMessage(messageID)
	if err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ------------------------------- #COMMENTARE IL MESSAGGIO# --------------------------
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1) Controllo che l'utente sia loggato
	isValid, userID, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Recupero gli ID della conversazione e del messaggio
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
		return
	}

	messageID, ok := getIntParam(w, ps, "message_id")
	if !ok {
		return
	}

	// 3) Verifico se il messaggio esiste
	messageExist, err := rt.db.DoesMessageExist(conversationID, messageID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !messageExist {
		http.Error(w, "Message Not Found", http.StatusNotFound)
		return
	}

	// 3) Verifico se la conversazione esiste
	chatType, err := rt.db.GetConversationType(conversationID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if chatType == "" {
		http.Error(w, "Conversation Not Found", http.StatusNotFound)
		return
	}

	// 4) Verifico se l'utente fa parte della conversazione
	isMember, err := rt.isUserInConversation(conversationID, userID, chatType)
	if err != nil {
		if err.Error() == "Invalid conversation type" {
			http.Error(w, "Invalid conversation type", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not a participant of this conversation", http.StatusNotFound)
		return
	}

	// 5) Recupero il corpo della richiesta
	var reaction CommentApi
	err = json.NewDecoder(r.Body).Decode(&reaction)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 6) controllo se il commento è valido
	if len(reaction.CommentEmoji) < 1 || len(reaction.CommentEmoji) > 20 {
		http.Error(w, "Invalid emoji code", http.StatusBadRequest)
		return
	}

	// 7) Aggiungo il commento al messaggio
	err = rt.db.AddCommentToMessage(messageID, userID, reaction.CommentEmoji)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// ------------------------------- #RIMUOVERE UNA REAZIONE# --------------------------
func (rt *_router) removeReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1) Controllo che l'utente sia loggato
	isValid, userID, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Recupero gli ID della conversazione e del messaggio
	conversationID, ok := getIntParam(w, ps, "conversation_id")
	if !ok {
		return
	}

	messageID, ok := getIntParam(w, ps, "message_id")
	if !ok {
		return
	}

	// 3) Verifico se il messaggio esiste
	messageExist, err := rt.db.DoesMessageExist(conversationID, messageID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !messageExist {
		http.Error(w, "Message Not Found", http.StatusNotFound)
		return
	}

	// 4) Verifico se la conversazione esiste
	chatType, err := rt.db.GetConversationType(conversationID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if chatType == "" {
		http.Error(w, "Conversation Not Found", http.StatusNotFound)
		return
	}

	// 5) Verifico se l'utente fa parte della conversazione
	isMember, err := rt.isUserInConversation(conversationID, userID, chatType)
	if err != nil {
		if err.Error() == "Invalid conversation type" {
			http.Error(w, "Invalid conversation type", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not a participant of this conversation", http.StatusNotFound)
		return
	}

	// 6) Verifico se l'utente ha reagito al messaggio
	hasReacted, err := rt.db.HasUserReactedToMessage(userID, messageID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !hasReacted {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return
	}

	// 7) Rimuovo la reazione
	err = rt.db.RemoveReactionByUser(userID, messageID)
	if err != nil {
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
