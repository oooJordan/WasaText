package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// ------------------ #CREAZIONE DI UNA NUOVA CONVERSAZIONE# ---------------------------
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
	// decodifico la richiesta
	var req ConversationRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ChatType.ChatType == "group_chat" && req.ImageGroup == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		host := r.Host
		req.ImageGroup = scheme + "://" + host + "/defaultimage/defaultGroup.png"
	}

	// controllo la richiesta se è valida
	if req.ChatType.ChatType == "" || req.StartMessage.Media == "" || len(req.Usersname) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	// converto la richiesta in un formato compatibile con il database
	dbReq := convertToDatabaseConversationRequest(req)

	// crea una nuova conversazione nel database
	conversationID, err := rt.db.CreateConversationDB(author, dbReq)
	if err != nil {
		ctx.Logger.Error(err)
		http.Error(w, "Error creating conversation", http.StatusInternalServerError)
		return
	}
	lastMessage := req.StartMessage
	// se successo, ritorna la risposta
	response := CreateConversationResponse{
		ConversationID: conversationID,
		Message:        "Conversation created successfully",
		LastMessage:    lastMessage,
		ImageGroup:     req.ImageGroup,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.Error(err) // Logga l'errore lato server
		http.Error(w, "Failed to encode response as JSON", http.StatusInternalServerError)
		return
	}
}

// ------------------------ #OTTENGO LA LISTA DELLE MIE CONVERSAZIONI# ------------------------------
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	isValid, user_id, err := rt.IsValidToken(r, w)
	if err != nil {
		// La risposta HTTP è già gestita all'interno di IsValidToken
		return
	}
	if !isValid {
		// Token non valido
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = rt.db.UpdateMessageDelivered(user_id)
	if err != nil {
		http.Error(w, "Failed to update message delivery", http.StatusInternalServerError)
		return
	}

	// recupero le conversazioni dell'utente dal database
	var convs []ConversationsApi
	my_conversations, err := rt.db.GetUserConversations(user_id)
	if err != nil {
		log.Println("Errore in GetUserConversations:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// converto le conversazioni nel formato API
	for i := 0; i < len(my_conversations); i++ {
		convs = append(convs, ConvertConversationFromDatabase(my_conversations[i]))
	}
	// struct che racchiude l'array in un object conversation
	rispo_conv := struct {
		Conversations []ConversationsApi `json:"conversation"`
	}{
		Conversations: convs,
	}
	// Conversione in JSON e invio della risposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rispo_conv); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

// ----------------------- #OTTENGO LA CRONOLOGIA DEI MESSAGGI# --------------------------
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
			Users []User `json:"users"`
		} `json:"utenti"`
		Messages []MessageRicvApi `json:"messages"`
	}{}

	// 9) Popolo con gli utenti che fanno parte della conversazione
	for _, dbUser := range dbUsers {
		userAPI := User{
			UserID:       dbUser.UserID,
			Nickname:     dbUser.Name,
			ProfileImage: dbUser.ProfileImage,
		}
		response.Utenti.Users = append(response.Utenti.Users, userAPI)
	}

	// 10) Popolo con i messaggi, i commenti e il read_status
	for _, msg := range dbMessages {
		// Converto i commenti dal DB all'API
		var commentArray []CommentApi
		for _, c := range msg.Comment {
			commentArray = append(commentArray, CommentApi{
				UserName:     c.UserName,
				CommentEmoji: c.CommentEmoji,
			})
		}

		// Converto il read_status dal DB all'API
		var readStatusArray []ReadStatusApi
		for _, rs := range msg.ReadStatus {
			var isReadInt, isDeliveredInt int
			if rs.IsRead {
				isReadInt = 1
			} else {
				isReadInt = 0
			}
			if rs.IsDelivered {
				isDeliveredInt = 1
			} else {
				isDeliveredInt = 0
			}
			readStatusArray = append(readStatusArray, ReadStatusApi{
				UserID:      rs.UserID,
				IsRead:      isReadInt,
				IsDelivered: isDeliveredInt,
			})
		}

		// Formattazione del messaggio per la risposta
		messageAPI := MessageRicvApi{
			UserName: msg.UserName,
			Message_ID: func() *int {
				if msg.MessageID.Valid {
					id := int(msg.MessageID.Int64)
					return &id
				}
				return nil
			}(),
			Testo:       msg.Testo,
			MessageType: msg.MessageType,
			Image:       msg.Image,
			Timestamp:   msg.Timestamp,
			IsForwarded: msg.IsForwarded,
			Comment:     commentArray,
			ReadStatus:  readStatusArray,
			ReplyToMessageID: func() *int {
				if msg.ReplyToMessageID.Valid {
					id := int(msg.ReplyToMessageID.Int64)
					return &id
				}
				return nil
			}(),
		}

		response.Messages = append(response.Messages, messageAPI)
	}

	// 11) Invio la risposta JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
