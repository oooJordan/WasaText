package api

//"time"

//"github.com/oooJordan/WasaText/service/database"

// Struttura del login
type LoginRequest struct {
	User string `json:"name"`
}

type LoginResponse struct {
	User_ID int `json:"user_id"`
}

// User rappresenta la struttura degli utenti restituiti
type UserList struct {
	Name         string `json:"name"`
	UserID       int    `json:"user_id"`
	ProfileImage string `json:"profile_image"`
}

/*

// Struttura per rappresentare un emoji nel commento
type CommentEmoji struct {
	CommentIdentifier int `json:"comment_identifier"`
}

// Struttura per rappresentare un singolo messaggio
type Message struct {
	Message_ID        int          `json:"message_id"`
	Sender            string       `json:"sender"`
	Timestamp         time.Time    `json:"timestamp"`
	ContentType       string       `json:"content_type"`        // testo, gif, gif+testo
	Content           string       `json:"content"`             // contenuto del messaggio
	ImageUrl          string       `json:"image_url,omitempty"` // URL dell'immagine (opzionale)
	StatusMessageRead bool         `json:"status_message_read"` // stato di lettura del messaggio
	Emoji             CommentEmoji `json:"emoji,omitempty"`     // emoji (opzionale)
}

// Struttura per rappresentare una conversazione che contiene i messaggi
type Conversation struct {
	Conversation_ID int       `json:"conversation_id"`
	Messages        []Message `json:"messages"` // lista di messaggi nella conversazione
}

// Metodi per convertire l'oggetto API in un oggetto compatibile con il database

// INVIO DATI UTENTE (ID E USERNAME)
func (details User) toDataBaseModel() database.User {
	return database.User{
		User_ID: details.User_ID,
	}
}


// INVIO ID DEL COMMENTO
func (comment CommentEmoji) toDataBaseModel() database.CommentEmoji {
	return database.CommentEmoji{
		CommentIdentifier: comment.CommentIdentifier,
	}
}

// INVIO MESSAGGIO
func (msg Message) toDataBaseModel() database.Message {
	return database.Message{
		Message_ID:        msg.Message_ID,
		Sender:            msg.Sender,
		Timestamp:         msg.Timestamp.Format(time.RFC3339), //converto `time.Time` in stringa
		ContentType:       msg.ContentType,
		Content:           msg.Content,
		ImageUrl:          msg.ImageUrl,
		StatusMessageRead: msg.StatusMessageRead,
		EmojiID:           msg.Emoji.CommentIdentifier,
	}
}

// INVIO CONVERSAZIONE
func (conversation Conversation) toDataBaseModel() database.Conversation {
	// Convertiamo i messaggi in un formato compatibile con il database
	var dbMessages []database.Message
	for _, msg := range conversation.Messages {
		dbMessages = append(dbMessages, msg.toDataBaseModel())
	}

	return database.Conversation{
		Conversation_ID: conversation.Conversation_ID,
		Messages:        dbMessages,
	}
}
*/
