package database

import (
	"database/sql"
	"errors"
)

// ------------------ #CREAZIONE DI UNA CONVERSAZIONE# --------------------------
func (db *appdbimpl) CreateConversationDB(author int, req ConversationRequest) (int, error) {
	trans, err := db.c.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := trans.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			return
		}
	}()

	var result sql.Result

	// Se è private_chat...
	if req.ChatType == "private_chat" {
		// 1) Deve avere esattamente 1 utente in req.Usersname
		if len(req.Usersname) != 1 {
			return 0, errors.New("private_chat must have exactly one other participant")
		}

		// 2) Controllo se esiste già una conversazione privata tra questi due utenti
		query := `
            SELECT conv.conversation_id
            FROM conversations conv
            INNER JOIN conversation_participants conv_p_1 ON conv.conversation_id = conv_p_1.conversation_id
            INNER JOIN conversation_participants conv_p_2 ON conv.conversation_id = conv_p_2.conversation_id
            WHERE conv.chatType = 'private_chat' AND conv_p_1.user_id = ? AND conv_p_2.user_id = ?;
        `
		var existingConvID int
		userID, errGet := db.GetUserIDByUsername(req.Usersname[0])
		if errGet != nil {
			return 0, errGet
		}

		err = db.c.QueryRow(query, author, userID).Scan(&existingConvID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, errors.New("error during creation chat")
			}
		} else {
			// Se non c’è errore, significa che esiste già
			return 0, errors.New("private conversation with these two users already exists")
		}

		// 3) Creo la conversazione come private_chat
		query = "INSERT INTO conversations (chatType) VALUES (?)"
		result, err = trans.Exec(query, req.ChatType)
		if err != nil {
			return 0, err
		}
	} else if req.ChatType == "group_chat" {
		if req.GroupName == "" {
			return 0, errors.New("groupName is required for group chat")
		}

		query := "INSERT INTO conversations (chatType, groupName, imageGroup) VALUES (?, ?, ?)"
		result, err = trans.Exec(query, req.ChatType, req.GroupName, req.ImageGroup)
		if err != nil {
			return 0, err
		}

	} else {
		return 0, errors.New("invalid chat type")
	}

	conversationID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.New("failed to retrieve last insert ID: " + err.Error())
	}

	// Controllo il messaggio iniziale
	switch req.StartMessage.Media {
	case "text":
		if req.StartMessage.Content == "" {
			return 0, errors.New("content is required for type 'text'")
		}
	case "gif":
		if req.StartMessage.Image == "" {
			return 0, errors.New("image is required for type 'gif'")
		}
	case "gif_with_text":
		if req.StartMessage.Content == "" || req.StartMessage.Image == "" {
			return 0, errors.New("both content and media are required for type 'gif_with_text'")
		}
	default:
		return 0, errors.New("invalid type: must be 'text', 'gif', or 'gif_with_text'")
	}

	// Inserisco il messaggio iniziale
	query := "INSERT INTO messages (conversation_id, user_id, content, media, type, is_forwarded) VALUES (?, ?, ?, ?, ?, ?)"
	result, err = trans.Exec(query, conversationID, author, req.StartMessage.Content, req.StartMessage.Image, req.StartMessage.Media, req.StartMessage.IsForwarded)
	if err != nil {
		return 0, errors.New("failed to insert start message: " + err.Error())
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.New("failed to retrieve last insert ID: " + err.Error())
	}

	// Aggiorno la tabella conversations con l'id del messaggio
	updateQuery := "UPDATE conversations SET message_id = ? WHERE conversation_id = ?"
	_, err = trans.Exec(updateQuery, messageID, conversationID)
	if err != nil {
		return 0, errors.New("failed to update message_id in conversation: " + err.Error())
	}

	// Aggiungo i partecipanti passati in req.Usersname
	for _, username := range req.Usersname {
		userID, errGet := db.GetUserIDByUsername(username)
		if errGet != nil {
			return 0, errGet
		}
		_, err = trans.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, userID)
		if err != nil {
			return 0, errors.New("failed to add member: " + username + " - " + err.Error())
		}
	}

	// Aggiungo l'autore stesso
	_, err = trans.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, author)
	if err != nil {
		return 0, errors.New("failed to add author to members: " + err.Error())
	}

	// Imposto messages_read_status per tutti gli utenti
	rows, err := trans.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, conversationID, author)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// Per gli utenti diversi dall'autore: is_delivered=0, is_read=0
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return 0, err
		}
		_, err = trans.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			messageID, userID)
		if err != nil {
			return 0, err
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// L’autore ha is_delivered=1 e is_read=1
	_, err = trans.Exec(
		`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
         VALUES (?, ?, 1, 1)`,
		messageID, author)
	if err != nil {
		return 0, err
	}

	// Commit finale
	err = trans.Commit()
	if err != nil {
		return 0, err
	}

	return int(conversationID), nil
}

// ------------------ #CONVERSAZIONI DI UN UTENTE# --------------------------
func (db *appdbimpl) GetUserConversations(author int) ([]Triplos, error) {
	query := `SELECT
					conversations.conversation_id,
					conversations.message_id,
					conversations.imageGroup,
					conversations.groupName,
					conversations.chatType,
					COALESCE(messages_read_status.is_read, FALSE) as is_read,
					COALESCE(messages_read_status.is_delivered, FALSE) AS is_delivered
				FROM
					conversations
				INNER JOIN 
					conversation_participants ON conversations.conversation_id = conversation_participants.conversation_id
				LEFT JOIN
					messages_read_status ON conversations.message_id = messages_read_status.message_id
					AND messages_read_status.user_id = ?
				LEFT JOIN
					messages ON conversations.message_id = messages.message_id
				WHERE
					conversation_participants.user_id = ?
				ORDER BY messages.timestamp DESC;
	`

	rows, err := db.c.Query(query, author, author)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("the user has not started any conversation")
		}
		return nil, errors.New("error executing query to fetch user conversations")
	}
	defer rows.Close()

	var conversations []Triplos

	for rows.Next() {
		var conv ConversationsDb
		var isRead bool
		var isDelivered bool
		if err := rows.Scan(&conv.ConversationId, &conv.MessageId, &conv.ChatImage, &conv.ChatName, &conv.ChatType, &isRead, &isDelivered); err != nil {
			return nil, errors.New("error scanning conversation row")
		}

		conv.MessageNotRead = isRead
		conv.MessageDelivered = isDelivered

		// Se è una chat privata, prendo i dati dell'altro utente
		if conv.ChatType == "private_chat" {
			q := `SELECT 
						users.name,
						users.profile_image
					FROM
						users
					INNER JOIN
						conversation_participants ON conversation_participants.user_id = users.user_id
					WHERE 
						conversation_participants.conversation_id = ? AND users.user_id != ?;`
			err := db.c.QueryRow(q, conv.ConversationId, author).Scan(&conv.ChatName, &conv.ChatImage)
			if err != nil {
				return nil, errors.New("error executing query to fetch user details")
			}
		}

		// Recupero l'ultimo messaggio solo se MessageId è valido (non NULL)
		var mex MessageRicvDb
		if conv.MessageId.Valid {
			qMex := `SELECT
							users.name,
							messages.timestamp,
							messages.type,
							messages.content,
							messages.media,
							messages.is_forwarded,
						FROM 
							messages 
						INNER JOIN
							users ON messages.user_id = users.user_id
						WHERE 
							message_id = ?;`
			err := db.c.QueryRow(qMex, conv.MessageId.Int64).Scan(&mex.UserName, &mex.Timestamp, &mex.MessageType, &mex.Testo, &mex.Image, &mex.IsForwarded)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					mex = MessageRicvDb{} // Messaggio non più esistente
				} else {
					return nil, errors.New("error executing query to fetch message details")
				}
			}
		} else {
			mex = MessageRicvDb{} // Nessun messaggio associato
		}

		// Recupero le reazioni (se ce ne sono)
		var comments []CommentDb
		if conv.MessageId.Valid {
			qComm := `SELECT
							reaction,
							user_id
						FROM 
							message_reactions 
						WHERE 
							message_id = ?;`
			rowComm, err := db.c.Query(qComm, conv.MessageId.Int64)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New("error executing query to fetch comment details")
			}
			if rowComm != nil {
				defer rowComm.Close()
				for rowComm.Next() {
					var commento CommentDb
					if err := rowComm.Scan(&commento.CommentEmoji, &commento.UserName); err != nil {
						return nil, errors.New("error scanning comments row")
					}
					comments = append(comments, commento)
				}
				if err := rowComm.Err(); err != nil {
					return nil, errors.New("error occurred during comments row iteration")
				}
			}
		}

		// Assembla il risultato
		var t Triplos
		t.Conversation = conv
		t.Message = mex
		t.Commento = comments
		t.ReadStatus = []ReadStatusDb{
			{
				UserID:      author,
				IsRead:      isRead,
				IsDelivered: isDelivered,
			},
		}
		conversations = append(conversations, t)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error occurred during conversations row iteration")
	}

	return conversations, nil
}
