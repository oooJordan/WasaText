package database

import (
	"database/sql"
	"errors"
)

// Funzione per ottenere le conversazioni di un utente
func (db *appdbimpl) GetUserConversations(author int) ([]Triplos, error) {
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
				conversation_participants.user_id = ?;
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
	var conversations []Triplos
	//itero sui risultati
	for rows.Next() {
		var conv ConversationsDb
		var mex MessageRicvDb
		var comments []CommentDb
		var c Triplos
		if err := rows.Scan(&conv.ConversationId, &conv.MessageId, &conv.ChatImage, &conv.ChatName, &conv.ChatType); err != nil {
			return nil, errors.New("error scanning conversation row")
		}

		if conv.ChatType == "private_chat" {
			query := `  SELECT 
								users.name,
								users.profile_image
						FROM
								conversation_participants
						INNER JOIN
								users ON conversation_participants.user_id = users.user_id
						WHERE 
								conversation_id = ? AND user_id != ?;
			`
			err := db.c.QueryRow(query, author).Scan(&conv.ChatName, &conv.ChatImage)
			if err != nil {
				return nil, errors.New("error executing query to fetch user details")
			}
		}
		query := ` SELECT
							users.name,
							messages.timestamp,
							messages.type,
							messages.content,
							messages.media
					FROM 
							messages 
					INNER JOIN
							users ON messages.user_id = users.user_id
					WHERE 
							message_id = ?; `

		err := db.c.QueryRow(query, conv.MessageId).Scan(&mex.UserID, &mex.Timestamp, &mex.MessageType, &mex.Testo, &mex.Image)
		if err != nil {
			return nil, errors.New("error executing query to fetch message details")
		}
		quer := ` SELECT
						reaction,
						user_id
					FROM 
						reactions 
					WHERE 
						message_id = ?;`

		rowComm, err := db.c.Query(quer, conv.MessageId)
		if err != nil {
			if err != sql.ErrNoRows {
				return nil, errors.New("error executing query to fetch comment details")
			}
		} else {
			defer rowComm.Close()

			for rowComm.Next() {
				var commento CommentDb
				if err := rowComm.Scan(&commento.CommentEmoji, &commento.UserID); err != nil {
					return nil, errors.New("error scanning comments row")
				}
				comments = append(comments, commento)
			}
		}
		//popolo i dati della tripla
		c.Conversation = conv
		c.Message = mex
		c.Commento = comments

		conversations = append(conversations, c)

	}

	return conversations, nil
}
