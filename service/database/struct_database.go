package database

// Struttura che rappresenta un utente nel database
type User struct {
	User_ID  int    `json:"user_id"`
	Username string `json:"username"`
}

// Struttura per rappresentare un emoji nel commento
type CommentEmoji struct {
	CommentIdentifier int `json:"comment_identifier"`
}

// Struttura per rappresentare un messaggio nel database
type Message struct {
	Message_ID        int    `json:"message_id"`
	Sender            string `json:"sender"`
	Timestamp         string `json:"timestamp"`
	ContentType       string `json:"content_type"`
	Content           string `json:"content"`
	ImageUrl          string `json:"image_url"`
	StatusMessageRead bool   `json:"status_message_read"`
	EmojiID           int    `json:"emoji_id"`
}

// Struttura per rappresentare una conversazione nel database
type Conversation struct {
	Conversation_ID int       `json:"conversation_id"`
	Messages        []Message `json:"messages"` // lista di messaggi
}
