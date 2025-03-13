package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) NewMessage(conversationID int, senderID int, messageType string, content string, media string) (int, error) {
	var messageID int

	// Inizio la transazione
	tx, err := db.c.Begin()
	if err != nil {
		return 0, err
	}

	// Inserisco il nuovo messaggio nella tabella messages
	err = tx.QueryRow(
		`INSERT INTO messages (conversation_id, user_id, type, content, media) 
         VALUES (?, ?, ?, ?, ?) RETURNING message_id`,
		conversationID, senderID, messageType, content, media,
	).Scan(&messageID) // Salvo l'ID del messaggio
	// Se c'è un errore, faccio il rollback della transazione
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Ottengo tutti gli utenti della conversazione (tranne il mittente)
	rows, err := tx.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, conversationID, senderID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer rows.Close() // Chiudo la query

	// Inserisco gli utenti nella tabella messages_read_status
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			tx.Rollback()
			return 0, err
		}
		// Inserisco lo stato di non consegnato e non letto per tutti gli utenti della conversazione
		_, err = tx.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			messageID, userID)
		if err != nil {
			tx.Rollback()
			return 0, err // Se c'è un errore, faccio il rollback della transazione
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

	// Commit della transazione (se tutto è andato bene)
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	// Ritorno l'ID del messaggio
	return messageID, nil
}

// Aggiorna lo stato di un messaggio a consegnato
func (db *appdbimpl) UpdateMessageDelivered(userID int) error {
	// Aggiorna lo stato di tutti i messaggi non consegnati dell'utente
	_, err := db.c.Exec(
		`UPDATE messages_read_status 
         SET is_delivered = TRUE 
         WHERE user_id = ? 
         AND is_delivered = FALSE`,
		userID)
	return err
}

// Aggiorna lo stato di un messaggio a letto
func (db *appdbimpl) UpdateMessageRead(userID int, conversationID int) error {
	// Aggiorna lo stato di tutti i messaggi non letti dell'utente
	_, err := db.c.Exec(
		`UPDATE messages_read_status 
         SET is_read = 1 
         WHERE user_id = ? AND message_id IN 
         (SELECT message_id FROM messages WHERE conversation_id = ?)`,
		userID, conversationID)
	return err
}

// recupera tutti i messaggi di una conversazione
func (db *appdbimpl) GetConversationMessages(conversationID int) ([]MessageFullDB, error) {
	// Query per recuperare tutti i messaggi di una conversazione
	rows, err := db.c.Query(`
        SELECT 
            m.message_id,
            u.name AS user_name,
            m.content,
            m.media,
            m."type",
            m.timestamp,
            c.name AS comment_user,
            cm.reaction
        FROM messages m
        JOIN users u ON m.user_id = u.user_id
        LEFT JOIN message_reactions cm ON m.message_id = cm.message_id
        LEFT JOIN users c ON cm.user_id = c.user_id
        WHERE m.conversation_id = ?
        ORDER BY m.timestamp ASC
    `, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Creo una mappa per tenere traccia dei messaggi
	results := []MessageFullDB{}
	msgIndexMap := make(map[int]int)
	// Scorro i risultati della query e li inserisco nella mappa dei messaggi (results)
	for rows.Next() {
		var (
			messageID       int
			userName        string
			content         string
			media           string
			MessageType     string
			timestamp       string
			commentUserName *string // *string per gestire il NULL
			commentEmoji    *string
		)
		err := rows.Scan(
			&messageID,
			&userName,
			&content,
			&media,
			&MessageType,
			&timestamp,
			&commentUserName,
			&commentEmoji,
		)
		// Se c'è un errore, lo ritorno
		if err != nil {
			return nil, err
		}
		// Se il messaggio non è già presente (!found), lo inserisco nella mappa dei messaggi
		idx, found := msgIndexMap[messageID]
		if !found {
			// Creo un nuovo messaggio e lo appendo alla lista dei messaggi
			newMsg := MessageFullDB{
				MessageID:   messageID,
				UserName:    userName,
				Testo:       content,
				MessageType: MessageType,
				Image:       media,
				Timestamp:   timestamp,
				Comment:     []CommentDb{},
			}
			results = append(results, newMsg)
			// Aggiorno l'indice dell'ultimo messaggio inserito nella mappa dei messaggi (results)
			idx = len(results) - 1
			msgIndexMap[messageID] = idx
		}

		// Se è presente un commento, lo aggiungo al messaggio
		if commentUserName != nil && commentEmoji != nil {
			c := CommentDb{
				UserName:     *commentUserName,
				CommentEmoji: *commentEmoji,
			}
			// Aggiungo il commento al messaggio
			results[idx].Comment = append(results[idx].Comment, c)
		}
	}
	// Controllo se ci sono errori
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Ritorno i messaggi e i commenti trovati nella query (results)
	return results, nil
}

func (db *appdbimpl) ForwardMessage(destinationConversationID int, originalMessageID int, forwardingUserID int) (int64, error) {
	// Verifica se l'utente ha scaricato e letto il messaggio originale
	var isDelivered, isRead bool
	query := `
		SELECT is_delivered, is_read 
		FROM messages_read_status 
		WHERE message_id = ? AND user_id = ?`
	err := db.c.QueryRow(query, originalMessageID, forwardingUserID).Scan(&isDelivered, &isRead)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("user has no read status for this message")
		}
		return 0, err
	}

	// Se l'utente non ha letto o scaricato il messaggio, non può inoltrarlo
	if !isDelivered || !isRead {
		return 0, errors.New("user must have downloaded and read the message before forwarding")
	}

	// Recupero il tipo, il contenuto e il media del messaggio originale
	var msgType, content, media string
	query = "SELECT type, content, media FROM messages WHERE message_id = ?"
	err = db.c.QueryRow(query, originalMessageID).Scan(&msgType, &content, &media)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("original message not found")
		}
		return 0, err
	}

	// Verifico che la conversazione di destinazione esista
	var convexist int
	err = db.c.QueryRow("SELECT conversation_id FROM conversations WHERE conversation_id = ?", destinationConversationID).Scan(&convexist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("destination conversation not found")
		}
		return 0, err
	}

	// Inizio una transazione per garantire la coerenza dei dati
	trans, err := db.c.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := trans.Rollback(); err != nil && err != sql.ErrTxDone {
			return
		}
	}()

	// Inserisco il messaggio inoltrato nella tabella messages
	insertQuery := `
		INSERT INTO messages (conversation_id, user_id, type, content, media)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := trans.Exec(insertQuery, destinationConversationID, forwardingUserID, msgType, content, media)
	if err != nil {
		return 0, err
	}

	// Ottengo l'ID del messaggio appena inserito
	newMessageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Aggiorno la tabella `messages_read_status`
	// Ottengo tutti gli utenti della conversazione (escludendo chi ha inoltrato il messaggio)
	rows, err := trans.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, destinationConversationID, forwardingUserID)
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
		// Imposto is_delivered = 0 e is_read = 0 per tutti gli utenti della conversazione (tranne chi inoltra il messaggio)
		_, err = trans.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			newMessageID, userID)
		if err != nil {
			trans.Rollback()
			return 0, err
		}
	}

	// Il mittente dell'inoltro ha sempre is_delivered = 1 e is_read = 1
	_, err = trans.Exec(
		`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
         VALUES (?, ?, 1, 1)`,
		newMessageID, forwardingUserID)
	if err != nil {
		trans.Rollback()
		return 0, err
	}

	// Commit della transazione
	err = trans.Commit()
	if err != nil {
		return 0, err
	}

	return newMessageID, nil
}
