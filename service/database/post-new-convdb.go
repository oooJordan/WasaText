package database

import (
	"database/sql"
	"fmt"
)

func (db *appdbimpl) CreateConversationDB(author int, req ConversationRequest) (int, error) {

	trans, err := db.c.Begin()
	if err != nil {
		return 0, err
	}
	defer trans.Rollback()

	var result sql.Result
	if req.ChatType == "private_chat" {
		query := "INSERT INTO conversations (chatType) VALUES (?)"
		result, err = trans.Exec(query, req.ChatType)
	} else if req.ChatType == "group_chat" {
		query := "INSERT INTO conversations (chatType, groupName, imageGroup) VALUES (?, ?, ?)"
		result, err = trans.Exec(query, req.ChatType, req.GroupName, req.ImageGroup)
	} else {
		return 0, fmt.Errorf("failed to create a conversation: %w", err)
	}

	// Recupero l'ID della nuova conversazione
	conversationID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	// Inserisco il messaggio iniziale
	query := "INSERT INTO messages (conversation_id, content, media, image) VALUES (?, ?, ?, ?)"
	_, err = trans.Exec(query, conversationID, req.StartMessage.Content, req.StartMessage.Media, req.StartMessage.Image)
	if err != nil {
		return 0, fmt.Errorf("error inserting start message: %w", err)
	}

	// Inserisco i membri (non me)
	for _, user := range req.Usersname {
		_, err := trans.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, user)
		if err != nil {
			return 0, fmt.Errorf("error adding member %v: %v", user, err)
		}
	}

	// Aggiungo me stessa come membro
	_, err = trans.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, author)
	if err != nil {
		return 0, fmt.Errorf("error adding author to members: %v", err)
	}

	// Inserisco il messaggio iniziale
	_, err = trans.Exec("INSERT INTO messages (conversation_id, content, media, image) VALUES (?, ?, ?, ?)", conversationID, req.StartMessage.Content, req.StartMessage.Media, req.StartMessage.Image)
	if err != nil {
		return 0, fmt.Errorf("error inserting start message: %v", err)
	}

	err = trans.Commit()
	if err != nil {
		return 0, err
	}

	return int(conversationID), nil

}
