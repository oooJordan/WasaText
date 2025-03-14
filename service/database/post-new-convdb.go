package database

import (
	"database/sql"
	"errors"
)

// #CREAZIONE DI UNA NUOVA CONVERSAZIONE#
// CreateConversationDB crea una nuova conversazione
func (db *appdbimpl) CreateConversationDB(author int, req ConversationRequest) (int, error) {
	DefaultImage := "https://cdn.raceroster.com/assets/images/team-placeholder.png"
	trans, err := db.c.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := trans.Rollback(); err != nil && err != sql.ErrTxDone {
			return
		}
	}()
	// controllo utenti chat privata
	if req.ChatType == "private_chat" {
		if len(req.Usersname) != 1 {
			return 0, errors.New("private_chat must have exactly one other participant")
		}
	}

	var result sql.Result
	if req.ChatType == "private_chat" {
		query := `
			SELECT conv.conversation_id
			FROM conversations conv
			INNER JOIN conversation_participants conv_p_1 ON conv.conversation_id = conv_p_1.conversation_id
			INNER JOIN conversation_participants conv_p_2 ON conv.conversation_id = conv_p_2.conversation_id
			WHERE conv.chatType = 'private_chat' AND conv_p_1.user_id = ? AND conv_p_2.user_id = ?;`

		var conversation_id int
		id_utente_2, _ := db.GetUserIDByUsername(req.Usersname[0])
		err = db.c.QueryRow(query, author, id_utente_2).Scan(&conversation_id)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, errors.New("error during creation chat")
			}
		} else {
			return 0, errors.New("private conversation with these two users already exists")
		}

		query = "INSERT INTO conversations (chatType) VALUES (?)"
		result, err = trans.Exec(query, req.ChatType)
	} else if req.ChatType == "group_chat" {
		if req.GroupName == "" {
			return 0, errors.New("groupName is required for group chat")
		}
		if req.ImageGroup == "" {
			req.ImageGroup = DefaultImage
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

	// Inserisco il messaggio dopo la validazione
	query := "INSERT INTO messages (conversation_id, user_id, content, media, type) VALUES (?, ?, ?, ?, ?)"
	result, err = trans.Exec(query, conversationID, author, req.StartMessage.Content, req.StartMessage.Image, req.StartMessage.Media)
	if err != nil {
		return 0, errors.New("failed to insert start message: " + err.Error())
	}

	// prendo id del messaggio appena creato
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

	// #MESSAGES_READ_STATUS#
	// Ottengo tutti gli utenti della conversazione (tranne il mittente)
	rows, err := trans.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, conversationID, author)
	if err != nil {
		trans.Rollback()
		return 0, err
	}
	defer rows.Close()

	// Inserisco gli utenti nella tabella messages_read_status
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			trans.Rollback()
			return 0, err
		}
		// Inserisco lo stato di non consegnato e non letto per tutti gli utenti della conversazione
		_, err = trans.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			messageID, userID)
		if err != nil {
			trans.Rollback()
			return 0, err
		}
	}

	// Il mittente ha sempre is_delivered = 1 e is_read = 1
	_, err = trans.Exec(
		`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
         VALUES (?, ?, 1, 1)`,
		messageID, author)
	if err != nil {
		trans.Rollback()
		return 0, err
	}

	// Commit finale della transazione
	err = trans.Commit()
	if err != nil {
		return 0, err
	}

	return int(conversationID), nil
}
