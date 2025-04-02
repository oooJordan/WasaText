package database

import (
	"database/sql"
	"errors"
	"log"
)

// ------------------------------- #INVIO DI UN NUOVO MESSAGGIO# --------------------------
// NewMessage crea un nuovo messaggio nella tabella messages e aggiorna la tabella messages_read_status
// per tutti gli utenti della conversazione
func (db *appdbimpl) NewMessage(conversationID int, senderID int, messageType string, content string, media string, replyTo *int) (int, error) {
	var messageID int

	// Inizio la transazione
	tx, err := db.c.Begin()
	if err != nil {
		return 0, err
	}

	// Funzione per gestire il rollback con log dell'errore
	rollback := func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Println("Errore nel rollback:", rollbackErr)
		}
	}

	// Inserisco il nuovo messaggio nella tabella messages e recupero il message_id
	err = tx.QueryRow(
		`INSERT INTO messages (conversation_id, user_id, type, content, media, reply_to_message_id) 
         VALUES (?, ?, ?, ?, ?, ?) RETURNING message_id`,
		conversationID, senderID, messageType, content, media, replyTo,
	).Scan(&messageID)
	if err != nil {
		rollback()
		return 0, err
	}

	// Aggiorno il campo message_id nella tabella conversations
	_, err = tx.Exec(`UPDATE conversations SET message_id = ? WHERE conversation_id = ?`, messageID, conversationID)
	if err != nil {
		rollback()
		return 0, err
	}

	// Inserisco lo stato di lettura/consegna per tutti gli altri partecipanti (escludendo il mittente)
	rows, err := tx.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, conversationID, senderID)
	if err != nil {
		rollback()
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			rollback()
			return 0, err
		}
		_, err = tx.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			messageID, userID)
		if err != nil {
			rollback()
			return 0, err
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// **Modifica principale:**
	// Inserisco anche il record per il mittente, impostando is_delivered e is_read a TRUE (1)
	_, err = tx.Exec(
		`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
         VALUES (?, ?, 1, 1)`,
		messageID, senderID)
	if err != nil {
		rollback()
		return 0, err
	}

	// Se tutto va bene, committo la transazione
	if err := tx.Commit(); err != nil {
		rollback()
		return 0, err
	}

	return messageID, nil
}

// ----------------------- #AGGIORNAMENTO STATO CONSEGNA MESSAGGIO# --------------------------
// UpdateMessageDelivered aggiorna lo stato di tutti i messaggi non consegnati dell'utente
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

// ------------------ #AGGIORNAMENTO STATO LETTURA MESSAGGIO# --------------------------
// UpdateMessageRead aggiorna lo stato di tutti i messaggi non letti dell'utente
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

// --------------------- #CRONOLOGIA MESSAGGI CONVERSAZIONE# --------------------------
// GetConversationMessages recupera tutti i messaggi di una conversazione
func (db *appdbimpl) GetConversationMessages(conversationID int) ([]MessageFullDB, error) {
	query := `
        SELECT 
            m.message_id,
            u.name AS user_name,
            m.content,
            m.media,
            m."type",
            m.timestamp,
			m.is_forwarded,
			m.reply_to_message_id,
            c.name AS comment_user,
            cm.reaction,
            mrs.user_id,
            mrs.is_read,
            mrs.is_delivered
        FROM messages m
        JOIN users u ON m.user_id = u.user_id
        LEFT JOIN message_reactions cm ON m.message_id = cm.message_id
        LEFT JOIN users c ON cm.user_id = c.user_id
        LEFT JOIN messages_read_status mrs ON m.message_id = mrs.message_id
        WHERE m.conversation_id = ?
        ORDER BY m.timestamp ASC
    `
	rows, err := db.c.Query(query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []MessageFullDB{}
	msgIndexMap := make(map[int64]int)

	for rows.Next() {
		var (
			messageID       sql.NullInt64
			userName        string
			content         string
			media           string
			messageType     string
			timestamp       string
			isForwarded     sql.NullBool
			replyTo         sql.NullInt64
			commentUserName *string
			commentEmoji    *string
			rsUserID        *int
			rsIsRead        sql.NullBool
			rsIsDelivered   sql.NullBool
		)

		err := rows.Scan(
			&messageID,
			&userName,
			&content,
			&media,
			&messageType,
			&timestamp,
			&isForwarded,
			&replyTo,
			&commentUserName,
			&commentEmoji,
			&rsUserID,
			&rsIsRead,
			&rsIsDelivered,
		)
		if err != nil {
			return nil, err
		}

		if !messageID.Valid {
			continue // se per qualche motivo è NULL, salta
		}

		id := messageID.Int64
		idx, found := msgIndexMap[id]
		if !found {
			newMsg := MessageFullDB{
				MessageID:        messageID,
				UserName:         userName,
				Testo:            content,
				MessageType:      messageType,
				Image:            media,
				Timestamp:        timestamp,
				IsForwarded:      isForwarded.Valid && isForwarded.Bool,
				ReplyToMessageID: replyTo,
				Comment:          []CommentDb{},
				ReadStatus:       []ReadStatusDb{},
			}
			results = append(results, newMsg)
			idx = len(results) - 1
			msgIndexMap[id] = idx
		}

		// Commenti
		if commentUserName != nil && commentEmoji != nil {
			alreadyExists := false
			for _, comm := range results[idx].Comment {
				if comm.UserName == *commentUserName && comm.CommentEmoji == *commentEmoji {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				results[idx].Comment = append(results[idx].Comment, CommentDb{
					UserName:     *commentUserName,
					CommentEmoji: *commentEmoji,
				})
			}
		}

		// Stato lettura/consegna
		if rsUserID != nil {
			var isRead, isDelivered bool
			if rsIsRead.Valid {
				isRead = rsIsRead.Bool
			}
			if rsIsDelivered.Valid {
				isDelivered = rsIsDelivered.Bool
			}

			alreadyExists := false
			for _, rs := range results[idx].ReadStatus {
				if rs.UserID == *rsUserID {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				results[idx].ReadStatus = append(results[idx].ReadStatus, ReadStatusDb{
					UserID:      *rsUserID,
					IsRead:      isRead,
					IsDelivered: isDelivered,
				})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// ------------------------------- #INOLTRO DI UN MESSAGGIO# --------------------------
// ForwardMessage inoltra un messaggio a una conversazione e aggiorna la tabella messages_read_status
// per tutti gli utenti della conversazione
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
		if err := trans.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			return
		}
	}()

	// Inserisco il messaggio inoltrato nella tabella messages
	insertQuery := `
		INSERT INTO messages (conversation_id, user_id, type, content, media, is_forwarded)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := trans.Exec(insertQuery, destinationConversationID, forwardingUserID, msgType, content, media, true)
	if err != nil {
		return 0, err
	}

	// Ottengo l'ID del messaggio appena inserito
	newMessageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Aggiorno il message_id nella tabella conversations
	_, err = trans.Exec(`UPDATE conversations SET message_id = ? WHERE conversation_id = ?`, newMessageID, destinationConversationID)
	if err != nil {
		if rollbackErr := trans.Rollback(); rollbackErr != nil {
			log.Printf("trans.Rollback() failed: %v", rollbackErr)
		}
		return 0, err
	}

	// Aggiorno la tabella `messages_read_status`
	// Ottengo tutti gli utenti della conversazione (escludendo chi ha inoltrato il messaggio)
	rows, err := trans.Query(`SELECT user_id FROM conversation_participants WHERE conversation_id = ? AND user_id != ?`, destinationConversationID, forwardingUserID)
	if err != nil {
		if rollbackErr := trans.Rollback(); rollbackErr != nil {
			log.Printf("trans.Rollback() failed: %v", rollbackErr)
		}
		return 0, err
	}

	defer rows.Close()

	// Inserisco gli utenti nella tabella messages_read_status
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			if rollbackErr := trans.Rollback(); rollbackErr != nil {
				log.Printf("trans.Rollback() failed: %v", rollbackErr)
			}
			return 0, err
		}

		// Imposto is_delivered = 0 e is_read = 0 per tutti gli utenti della conversazione (tranne chi inoltra il messaggio)
		_, err = trans.Exec(
			`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
             VALUES (?, ?, 0, 0)`,
			newMessageID, userID)
		if err != nil {
			if rollbackErr := trans.Rollback(); rollbackErr != nil {
				log.Printf("trans.Rollback() failed: %v", rollbackErr)
			}
			return 0, err
		}

	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Il mittente dell'inoltro ha sempre is_delivered = 1 e is_read = 1
	_, err = trans.Exec(
		`INSERT INTO messages_read_status (message_id, user_id, is_delivered, is_read)
         VALUES (?, ?, 1, 1)`,
		newMessageID, forwardingUserID)
	if err != nil {
		if rollbackErr := trans.Rollback(); rollbackErr != nil {
			log.Printf("trans.Rollback() failed: %v", rollbackErr)
		}
		return 0, err
	}

	// Commit della transazione
	err = trans.Commit()
	if err != nil {
		return 0, err
	}

	return newMessageID, nil
}

// ------------------------- #MITTENTE DEL MESSAGGIO# --------------------------
// GetMessageSender recupera l'ID del mittente di un messaggio
func (db *appdbimpl) GetMessageSender(messageID int, conversationID int) (int, error) {
	var senderID int
	err := db.c.QueryRow(`SELECT user_id FROM messages WHERE message_id = ? AND conversation_id = ?`, messageID, conversationID).Scan(&senderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("message not found")
		}
		return 0, err
	}
	return senderID, nil
}

// ---------------- #ELIMINAZIONE DEL MESSAGGIO IN MESSAGES# --------------------------
// DeleteMessage elimina il messaggio dalla tabella messages
func (db *appdbimpl) DeleteMessage(messageID int, conversationID int) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}

	// Funzione per eseguire il rollback e gestire errori di rollback
	rollback := func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Println("Errore nel rollback:", rollbackErr)
		}
	}

	// 1. Elimino il messaggio dalla tabella messages
	_, err = tx.Exec(`DELETE FROM messages WHERE message_id = ?`, messageID)
	if err != nil {
		rollback()
		return err
	}

	// 2. Trovo il messaggio più recente ancora presente nella conversazione
	var newLastMessageID sql.NullInt64
	err = tx.QueryRow(`SELECT MAX(message_id) FROM messages WHERE conversation_id = ?`, conversationID).Scan(&newLastMessageID)
	if err != nil {
		rollback()
		return err
	}

	// 3. Aggiorno message_id in conversations
	_, err = tx.Exec(`UPDATE conversations SET message_id = ? WHERE conversation_id = ?`, newLastMessageID, conversationID)
	if err != nil {
		rollback()
		return err
	}

	// 4. Commit della transazione
	if err := tx.Commit(); err != nil {
		rollback()
		return err
	}

	return nil
}

// --------------- #RIMUOVERE IL MESSAGGIO IN MESSAGES_READ_STATUS# ---------------------
// DeleteMessageStatus elimina il messaggio dalla tabella messages_read_status
func (db *appdbimpl) DeleteMessageStatus(messageID int) error {
	_, err := db.c.Exec(`DELETE FROM messages_read_status WHERE message_id = ?`, messageID)
	return err
}

// ------------------- #AGGIUNGE UNA REAZIONE AD UN MESSAGGIO# --------------------------
func (db *appdbimpl) AddCommentToMessage(messageID int, userID int, reaction string) error {
	// controllo se l'utente ha già commentato il messaggio
	var existReaction string
	err := db.c.QueryRow(
		`SELECT reaction FROM message_reactions WHERE message_id = ? AND user_id = ?`, messageID, userID).Scan(&existReaction)

	if errors.Is(err, sql.ErrNoRows) {
		// se non esiste, inserisco il commento
		_, err = db.c.Exec(
			`INSERT INTO message_reactions (message_id, user_id, reaction) VALUES (?, ?, ?)`,
			messageID, userID, reaction)
	} else if err == nil {
		// se esiste, aggiorno il commento
		_, err = db.c.Exec(
			`UPDATE message_reactions SET reaction = ? WHERE message_id = ? AND user_id = ?`,
			reaction, messageID, userID)
	}

	return err
}

// -------------------- #RIMUOVERE UNA REACTION DA UN MESSAGGIO# --------------------------
func (db *appdbimpl) RemoveReactionByUser(userID int, messageID int) error {
	_, err := db.c.Exec(`
        DELETE FROM message_reactions 
        WHERE message_id = ? AND user_id = ?`,
		messageID, userID)
	return err
}

func (db *appdbimpl) RemoveAllReaction(messageID int) error {
	_, err := db.c.Exec(`
        DELETE FROM message_reactions 
        WHERE message_id = ?`,
		messageID)
	return err
}
