package database

import (
	"database/sql"
	"errors"
)

// Funzione per ottenere le conversazioni di un utente
func (db *appdbimpl) GetUserConversations(author int) ([]ConversationsDb, error) {
	query := ` SELECT
				conversations.conversation_id,
				conversations.chatType,
				conversations.groupName,
				conversations.imageGroup,
				conversations.message_id
			FROM
				conversation_participants
			NATURAL JOIN 
				conversations
			WHERE
				conversation_participants.user_id = ?
	`
	rows, err := db.c.Query(query, author)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.New("the user has not started any conversation")
		}
	} else {
		return nil, errors.New("error executing query to fetch user conversations")
	}
	defer rows.Close()
	var conv ConversationsDb
	//itero sui risultati
	for rows.Next() {
		if err := rows.Scan(&conv.ConversationId, &conv.MessageId, &conv.ImageGroup, &conv.GroupName, &conv.ChatType); err != nil {
			return nil, errors.New("error scanning conversation row")
		}

	}

	return []ConversationsDb{}, nil
}
