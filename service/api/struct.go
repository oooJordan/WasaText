package api

import (
	"time"

	"github.com/oooJordan/WasaText/service/database"
)

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

type UpdateProfileImageRequest struct {
	Image string `json:"image"`
}

// User rappresenta la struttura degli utenti restituiti
type UserList struct {
	Name         string `json:"name"`
	UserID       int    `json:"user_id"`
	ProfileImage string `json:"profile_image"`
}

type ConversationRequest struct {
	ChatType     ResChatType `json:"chatType"`
	ImageGroup   string      `json:"imageGroup,omitempty"`
	GroupName    string      `json:"groupName,omitempty"`
	Usersname    []string    `json:"usersname"`
	StartMessage MessageSent `json:"startMessage"`
}

type ResChatType struct {
	ChatType string `json:"ChatType"`
}

type CreateConversationResponse struct {
	ConversationID int         `json:"ConversationId"`
	Message        string      `json:"message"`
	LastMessage    MessageSent `json:"lastMessage"`
}

type MessageSent struct {
	Content string `json:"content,omitempty"`
	Media   string `json:"media"`
	Image   string `json:"image,omitempty"`
}

type Conversations []struct {
	ChatType       string      `json:"chatType"`
	ConversationID int         `json:"conversationId"`
	UserName       string      `json:"nameChat,omitempty"`
	ImageGroup     string      `json:"imageGroup,omitempty"`
	GroupName      string      `json:"groupName,omitempty"`
	UserImage      string      `json:"profileimage,omitempty"`
	LastMessage    MessageSent `json:"lastMessage:"`
	Timestamp      string      `json:"timestamp"`
	Isread         bool        `json:"is_read"`
}

func convertToDatabaseConversationRequest(req ConversationRequest) database.ConversationRequest {
	return database.ConversationRequest{
		ChatType:   req.ChatType.ChatType,
		ImageGroup: req.ImageGroup,
		GroupName:  req.GroupName,
		Usersname:  req.Usersname,
		StartMessage: database.MessageSent{
			Content: req.StartMessage.Content,
			Media:   req.StartMessage.Media,
			Image:   req.StartMessage.Image,
		},
	}
}

type ConversationsApi struct {
	ConversationID int            `json:"conversationId"`
	Message        MessageRicvApi `json:"lastMessage"`
	ChatImage      string         `json:"profileimage"`
	ChatName       string         `json:"nameChat"`
	ChatType       string         `json:"ChatType"`
	MessageNotRead bool           `json:"statusMessageRead"`
}

type MessageRicvApi struct {
	UserName    string       `json:"username"`
	Message_ID  int          `json:"message_id"`
	Testo       string       `json:"content"`
	MessageType string       `json:"media"`
	Image       string       `json:"image"`
	Timestamp   string       `json:"timestamp"`
	Comment     []CommentApi `json:"comments"`
}

type CommentApi struct {
	UserName     string `json:"username"`
	CommentEmoji string `json:"emojiCode"`
}

func ConvertConversationFromDatabase(req database.Triplos) ConversationsApi {
	// converto i commenti dal database alla struttura API
	var comments []CommentApi
	for _, comment := range req.Commento {
		comments = append(comments, CommentApi{
			UserName:     comment.UserName,
			CommentEmoji: comment.CommentEmoji,
		})
	}

	// converto il messaggio
	message := MessageRicvApi{
		UserName:    req.Message.UserName,
		Timestamp:   req.Message.Timestamp.Time.Format(time.RFC3339),
		MessageType: req.Message.MessageType,
		Testo:       req.Message.Testo,
		Image:       req.Message.Image,
		Message_ID:  req.Conversation.MessageId,
		Comment:     comments,
	}
	// converto la conversazione
	return ConversationsApi{
		ConversationID: req.Conversation.ConversationId,
		Message:        message,
		ChatType:       req.Conversation.ChatType,
		ChatName:       req.Conversation.ChatName.String,
		ChatImage:      req.Conversation.ChatImage.String,
		MessageNotRead: req.Conversation.MessageNotRead,
	}
}
