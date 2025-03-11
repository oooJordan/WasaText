package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) AddUserToGroup(conversationID int, userID int) error {
	// Controllo se l'utente è già nel gruppo
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?", conversationID, userID).Scan(&count)
	if err != nil {
		return err
	}

	// Se l'utente è già nel gruppo, ritorno un errore
	if count > 0 {
		return errors.New("user already in group")
	}

	// Se l'utente non è nel gruppo, lo inserisco
	_, err = db.c.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, userID)
	return err
}

func (db *appdbimpl) RemoveUserFromGroup(conversationID int, userID int) error {
	result, err := db.c.Exec("DELETE FROM conversation_participants WHERE conversation_id = ? AND user_id = ?", conversationID, userID)
	if err != nil {
		return err
	}
	// verifico se l'eliminazione ha avuto effetto (cioè se l'utente era presente nel gruppo)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows // utente non trovato nel gruppo
	}
	return nil // operazione andata a buon fine
}

func (db *appdbimpl) UpdateGroupName(conversationID int, newName string) error {
	_, err := db.c.Exec(
		"UPDATE conversations SET groupName = ? WHERE conversation_id = ?",
		newName, conversationID,
	)
	return err
}

func (db *appdbimpl) DeleteGroup(conversationID int) error {
	_, err := db.c.Exec("DELETE FROM conversations WHERE conversation_id = ?", conversationID)
	return err
}

func (db *appdbimpl) UpdateGroupImage(conversationID int, Image string) error {
	_, err := db.c.Exec(
		"UPDATE conversations SET imageGroup = ? WHERE conversation_id = ?",
		Image, conversationID,
	)
	return err
}
