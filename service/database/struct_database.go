package database

import "database/sql"

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
	Media   string `json:"media"`
	Image   string `json:"image,omitempty"`
}

// GET CONVERSATIONS
type Triplos struct {
	Conversation ConversationsDb
	Message      MessageRicvDb
	Commento     []CommentDb
	ReadStatus   []ReadStatusDb
}

type ConversationsDb struct {
	ConversationId   int
	MessageId        sql.NullInt64
	ChatImage        sql.NullString
	ChatName         sql.NullString
	ChatType         string
	MessageNotRead   bool
	MessageDelivered bool
}

type MessageRicvDb struct {
	UserName    string
	Timestamp   sql.NullTime
	MessageType string
	Testo       string
	Image       string
}

type CommentDb struct {
	UserName     string
	CommentEmoji string
}

type ReadStatusDb struct {
	UserID      int
	IsRead      bool
	IsDelivered bool
}

type MessageFullDB struct {
	MessageID   sql.NullInt64
	UserName    string
	Testo       string
	MessageType string
	Image       string
	Timestamp   string
	Comment     []CommentDb
	ReadStatus  []ReadStatusDb
}
