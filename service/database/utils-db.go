package database

import "database/sql"

func (db *appdbimpl) CheckIDDatabase(userid int) (bool, error) {
	var found bool
	query := "SELECT 1 FROM users WHERE user_id = ?"
	err := db.c.QueryRow(query, userid).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Nessun utente trovato
		}
		return false, err // C'è stato un errore
	}
	return true, nil // Utente trovato
}
