package database

func (db *appdbimpl) NewMessage(conversationID int, senderID int, messageType string, content string, media string) (int, error) {
	var messageID int
	err := db.c.QueryRow(
		`INSERT INTO messages (conversation_id, user_id, type, content, media) 
         VALUES (?, ?, ?, ?, ?) RETURNING message_id`,
		conversationID, senderID, messageType, content, media,
	).Scan(&messageID)
	if err != nil {
		return 0, err
	}
	return messageID, nil
}
