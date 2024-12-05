package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) CreateConversationDB(author int, req ConversationRequest) (int, error) {
	DefaultImage := "https://cdn.raceroster.com/assets/images/team-placeholder.png"
	trans, err := db.c.Begin()
	if err != nil {
		return 0, err
	}
	defer trans.Rollback()
	// controllo utenti chat privata
	if req.ChatType == "private_chat" {
		if len(req.Usersname) != 1 {
			return 0, errors.New("private_chat must have exactly one other participant")
		}
	}

	var result sql.Result
	if req.ChatType == "private_chat" {
		query := "INSERT INTO conversations (chatType) VALUES (?)"
		result, err = trans.Exec(query, req.ChatType)
	} else if req.ChatType == "group_chat" {
		if req.GroupName == "" || req.ImageGroup == "" {
			if req.GroupName == "" {
				return 0, errors.New("groupName and imageGroup are required for group chat")
			}
			if req.ImageGroup == "" {
				req.ImageGroup = DefaultImage
			}

		}
		query := "INSERT INTO conversations (chatType, groupName, imageGroup) VALUES (?, ?, ?)"
		result, err = trans.Exec(query, req.ChatType, req.GroupName, req.ImageGroup)
	} else {
		return 0, errors.New("invalid chat type")
	}
	if err != nil {
		return 0, err
	}

	// ID della nuova conversazione
	conversationID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.New("failed to retrieve last insert ID: " + err.Error())
	}

	// controllo il messaggio
	switch req.StartMessage.Type {
	case "text":
		if req.StartMessage.Content == "" {
			return 0, errors.New("content is required for type 'text'")
		}
		if req.StartMessage.Media != "" {
			return 0, errors.New("media must be empty for type 'text'")
		}
	case "image":
		if req.StartMessage.Media == "" {
			return 0, errors.New("media is required for type 'image'")
		}
		if req.StartMessage.Content != "" {
			return 0, errors.New("content must be empty for type 'image'")
		}
	case "text_image":
		if req.StartMessage.Content == "" || req.StartMessage.Media == "" {
			return 0, errors.New("both content and media are required for type 'text_image'")
		}
	default:
		return 0, errors.New("invalid type: must be 'text', 'image', or 'text_image'")
	}

	// Inserisco il messaggio dopo la validazione
	query := "INSERT INTO messages (conversation_id, user_id, content, media, type) VALUES (?, ?, ?, ?, ?)"
	_, err = trans.Exec(query, conversationID, author, req.StartMessage.Content, req.StartMessage.Media, req.StartMessage.Type)
	if err != nil {
		return 0, errors.New("failed to insert start message: " + err.Error())
	}

	// prima recupero id dell'utente partendo dal suo username (per ogni utente)
	for _, username := range req.Usersname {
		userID, err := db.GetUserIDByUsername(username)
		if err != nil {
			return 0, err
		}

		// una volta preso l'id li aggiungo alla conversazione
		_, err = trans.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, userID)
		if err != nil {
			return 0, errors.New("failed to add member: " + username + " - " + err.Error())
		}
	}

	// aggiungo l'autore tra i partecipanti
	_, err = trans.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, author)
	if err != nil {
		return 0, errors.New("failed to add author to members: " + err.Error())
	}

	err = trans.Commit()
	if err != nil {
		return 0, err
	}

	return int(conversationID), nil
}
