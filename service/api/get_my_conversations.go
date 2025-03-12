package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

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
