package database

// Struttura che rappresenta un utente nel database
type UserIdDatabase struct {
	User_ID int `json:"user_id"`
}

// Struttura Users
type Users struct {
	Name         string `json:"name"`
	UserID       int    `json:"user_id"`
	ProfileImage string `json:"profile_image"`
}

// User-token struct
type UserToken struct {
	UserID int    `json:"user-id"`
	Token  string `json:"auth-token"`
}

type ConversationRequest struct {
	ChatType     string      `json:"chatType"`
	ImageGroup   string      `json:"imageGroup,omitempty"`
	GroupName    string      `json:"groupName,omitempty"`
	Usersname    []string    `json:"usersname"`
	StartMessage MessageSent `json:"startMessage"`
}

type MessageSent struct {
	Content string `json:"content,omitempty"`
	Media   string `json:"media,omitempty"`
	Type    string `json:"type"`
}

type ConversationsDb struct {
	ConversationId int
	Message        MessageRicvDb
	ProfileImage   string
	ChatName       string
	ChatType       string
}

type MessageRicvDb struct {
	UserName          string
	Message_ID        int
	Text_message      string
	Type_message      string
	Image             string
	Timestamp         string
	StatusMessageRead bool
	Comment           []CommentDb
}

type CommentDb struct {
	UserName     string
	CommentEmoji string
}

/*
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
*/
