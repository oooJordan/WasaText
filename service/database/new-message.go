package database

func (db *appdbimpl) NewMessage(conversationID int, senderID int, messageType string, content string, media string) (int, error) {
	var messageID int

	// Inizia la transazione
	tx, err := db.c.Begin()
	if err != nil {
		return 0, err
	}

	// Inserisci il nuovo messaggio nella tabella messages
	err = tx.QueryRow(
		`INSERT INTO messages (conversation_id, user_id, type, content, media) 
         VALUES (?, ?, ?, ?, ?) RETURNING message_id`,
		conversationID, senderID, messageType, content, media,
	).Scan(&messageID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Ottenere tutti gli utenti della conversazione (tranne il mittente)
	rows, err := tx.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, conversationID, senderID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer rows.Close()

	// Inserisci gli utenti nella tabella messages_read_status
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			tx.Rollback()
			return 0, err
		}

		_, err = tx.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			messageID, userID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Il mittente ha sempre is_delivered = 1 e is_read = 1
	_, err = tx.Exec(
		`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
         VALUES (?, ?, 1, 1)`,
		messageID, senderID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit della transazione
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return messageID, nil
}

func (db *appdbimpl) UpdateMessageDelivered(userID int) error {
	_, err := db.c.Exec(
		`UPDATE messages_read_status 
         SET is_delivered = TRUE 
         WHERE user_id = ? 
         AND is_delivered = FALSE`,
		userID)
	return err
}

func (db *appdbimpl) UpdateMessageRead(userID int, conversationID int) error {
	_, err := db.c.Exec(
		`UPDATE messages_read_status 
         SET is_read = 1 
         WHERE user_id = ? AND message_id IN 
         (SELECT message_id FROM messages WHERE conversation_id = ?)`,
		userID, conversationID)
	return err
}
