package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) CheckIDDatabase(userid int) (bool, string, error) {
	var name string

	// Query per verificare se l'utente esiste e ottenere il nome utente
	query := "SELECT name FROM users WHERE user_id = ?"
	err := db.c.QueryRow(query, userid).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", nil // Nessun utente trovato
		}
		return false, "", err // C'è stato un errore
	}

	// Se l'utente è stato trovato, restituiamo true e l'username
	return true, name, nil
}

// ritorna l'userID partendo dall'username di un utente
func (db *appdbimpl) GetUserIDByUsername(username string) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT user_id FROM users WHERE name = ?", username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("user not found: " + username)
		}
		return 0, errors.New("failed to retrieve user ID: " + err.Error())
	}
	return userID, nil
}

func (db *appdbimpl) IsGroupConversation(conversationID int) (bool, error) {
	var chatType string
	err := db.c.QueryRow("SELECT chatType FROM conversations WHERE conversation_id = ?", conversationID).Scan(&chatType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // conversazione non trovata
		}
		return false, err
	}
	return chatType == "group_chat", nil
}

func (db *appdbimpl) IsGroupEmpty(conversationID int) (bool, error) {
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ?", conversationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (db *appdbimpl) IsUserInGroup(conversationID int, userID int) (bool, error) {
	var count int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		conversationID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *appdbimpl) GetConversationType(conversationID int) (string, error) {
	var chatType string
	err := db.c.QueryRow(
		"SELECT chatType FROM conversations WHERE conversation_id = ?",
		conversationID,
	).Scan(&chatType)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Conversazione non trovata
		}
		return "", err
	}
	return chatType, nil
}

func (db *appdbimpl) IsUserInPrivateChat(conversationID int, userID int) (bool, error) {
	var count int
	err := db.c.QueryRow(
		`SELECT COUNT(*) 
         FROM conversation_participants 
         WHERE conversation_id = ? AND user_id = ?`,
		conversationID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
