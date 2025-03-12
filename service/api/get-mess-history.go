package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

func (rt *_router) messageHistory(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1) Controllo se il token è valido
	isValid, userID, err := rt.IsValidToken(r, w)
	if err != nil || !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Recupero conversationID
	conversationID, err := strconv.Atoi(ps.ByName("conversation_id"))
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
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
	var isMember bool
	if chatType == "group_chat" {
		isMember, err = rt.db.IsUserInGroup(conversationID, userID)
	} else {
		isMember, err = rt.db.IsUserInPrivateChat(conversationID, userID)
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// 5) Aggiorno lo stato dei messaggi letti
	err = rt.db.UpdateMessageRead(userID, conversationID)
	if err != nil {
		http.Error(w, "Failed to update message read status", http.StatusInternalServerError)
		return
	}

	// 6) Recupero la lista degli utenti
	dbUsers, err := rt.db.GetConversationUsers(conversationID)
	if err != nil {
		http.Error(w, "Failed to fetch conversation users", http.StatusInternalServerError)
		return
	}

	// 7) Recupero tutti i messaggi + i commenti
	dbMessages, err := rt.db.GetConversationMessages(conversationID)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	// 8) Creo la struttura per contenere la risposta
	response := struct {
		Utenti struct {
			Users []UserList `json:"users"`
		} `json:"utenti"`
		Messages []struct {
			Message MessageRicvApi `json:"message"`
		} `json:"messages"`
	}{}

	// 9) Popolo con gli utenti che fanno parte della conversazione
	for _, dbUser := range dbUsers {
		userAPI := UserList{
			UserID:       dbUser.UserID,
			Name:         dbUser.Name,
			ProfileImage: dbUser.ProfileImage,
		}
		// Aggiungo l'utente alla risposta
		response.Utenti.Users = append(response.Utenti.Users, userAPI)
	}

	// 10) Popolo con i messaggi e i commenti
	for _, msg := range dbMessages {
		// Converto i commenti dal DB all'API
		var commentArray []CommentApi
		for _, c := range msg.Comment {
			// Aggiungo i commenti all'array
			commentArray = append(commentArray, CommentApi{
				UserName:     c.UserName,
				CommentEmoji: c.CommentEmoji,
			})
		}

		// Formattazione del messaggio per la risposta
		messageAPI := MessageRicvApi{
			UserName:    msg.UserName,
			Message_ID:  msg.MessageID,
			Testo:       msg.Testo,
			MessageType: msg.MessageType,
			Image:       msg.Image,
			Timestamp:   msg.Timestamp,
			Comment:     commentArray,
		}
		// Aggiungo il messaggio alla risposta
		response.Messages = append(response.Messages, struct {
			Message MessageRicvApi `json:"message"`
		}{
			Message: messageAPI,
		})
	}

	// 11) Invio la risposta JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
