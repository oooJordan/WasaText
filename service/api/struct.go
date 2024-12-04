package api

import "github.com/oooJordan/WasaText/service/database"

//"time"

// Struttura del login
type LoginRequest struct {
	User string `json:"name"`
}

type LoginResponse struct {
	User_ID int `json:"user_id"`
}

// Struttura per mappare il corpo JSON
type UpdateUsernameRequest struct {
	NewUsername string `json:"newUsername"`
}

// User rappresenta la struttura degli utenti restituiti
type UserList struct {
	Name         string `json:"name"`
	UserID       int    `json:"user_id"`
	ProfileImage string `json:"profile_image"`
}

type ConversationRequest struct {
	ChatType     string      `json:"chatType"`
	ImageGroup   string      `json:"imageGroup,omitempty"`
	GroupName    string      `json:"groupName,omitempty"`
	Usersname    []string    `json:"usersname"`
	StartMessage MessageSent `json:"startMessage"`
}

type CreateConversationResponse struct {
	ConversationID int         `json:"ConversationId"`
	Message        string      `json:"message"`
	LastMessage    MessageSent `json:"lastMessage"`
}

type MessageSent struct {
	Content string `json:"content,omitempty"`
	Media   string `json:"media,omitempty"`
	Type    string `json:"type"`
}

func convertToDatabaseConversationRequest(req ConversationRequest) database.ConversationRequest {
	return database.ConversationRequest{
		ChatType:   req.ChatType,
		ImageGroup: req.ImageGroup,
		GroupName:  req.GroupName,
		Usersname:  req.Usersname,
		StartMessage: database.MessageSent{
			Content: req.StartMessage.Content,
			Media:   req.StartMessage.Media,
			Type:    req.StartMessage.Type,
		},
	}
}
