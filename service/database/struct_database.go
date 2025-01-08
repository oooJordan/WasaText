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
	Media   string `json:"media,omitempty"`
	Type    string `json:"type"`
}

// GET CONVERSATIONS
type Triplos struct {
	Conversation ConversationsDb
	Message      MessageRicvDb
	Commento     []CommentDb
}

type ConversationsDb struct {
	ConversationId int
	MessageId      int
	ChatImage      string
	ChatName       string
	ChatType       string
	//MessageNotRead int
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
	MessageId    int
	CommentEmoji string
}
