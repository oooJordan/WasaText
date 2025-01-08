package api

import "github.com/oooJordan/WasaText/service/database"

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

type ConversationsApi struct {
	ConversationID int            `json:"conversationId"`
	Message        MessageRicvApi `json:"lastMessage"`
	ChatImage      string         `json:"profileimage"`
	ChatName       string         `json:"nameChat"`
	ChatType       string         `json:"ChatType"`
	MessageNotRead int            `json:"statusMessageRead"`
}

type MessageRicvApi struct {
	UserName      string          `json:"username"`
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
			UserID:       comment.UserName,
			CommentEmoji: comment.CommentEmoji,
		})
	}

	// converto il messaggio
	message := MessageRicvApi{
		UserName:      req.Message.UserName,
		Timestamp:   req.Message.Timestamp.Time.GoString(),
		MessageType: req.Message.MessageType,
		Content:     req.Message.Testo,
		Image:       req.Message.Image,
		Comments:    comments,
	}
	return ConversationsApi{
		ConversationID: req.Conversation.ConversationId,
		ChatType:       req.Conversation.ChatType,
		ChatName:       req.Conversation.ChatName,
		ChatImage:      req.Conversation.ChatImage,
		MessageNotRead: req.Conversation.MessageNotRead,
		Message:        message,
	}
}

/*

func convertCommentFromDatabase(req []database.CommentDb) []CommentApi {
	comments := make([]CommentApi, len(req))
	for i, comment := range req {
		comments[i] = CommentApi{
			UserName:     comment.UserName,
			CommentEmoji: comment.CommentEmoji,
		}
	}
	return comments
}

func convertMessageFromDatabase(req database.MessageRicvDb) MessageRicvApi {
	return MessageRicvApi{
		UserName:     req.UserName,
		Message_ID:   req.Message_ID,
		Text_message: req.Text_message,
		Type_message: req.Type_message,
		Image:        req.Image,
		Timestamp:    req.Timestamp,
		Comment:      convertCommentFromDatabase(req.Comment),
	}

}

func convertConversationFromDatabase(req database.ConversationsDb) ConversationsApi {
	return ConversationsApi{
		ConversationId: req.ConversationId,
		Message:        convertMessageFromDatabase(req.Message),
		ProfileImage:   req.ProfileImage,
		ChatName:       req.ChatName,
		ChatType:       req.ChatType,
	}
}
*/
