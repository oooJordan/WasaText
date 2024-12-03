package database

import "database/sql"

func (db *appdbimpl) CheckIDDatabase(userid int) (bool, string, error) {
	var username string

	// Query per verificare se l'utente esiste e ottenere il nome utente
	query := "SELECT username FROM users WHERE user_id = ?"
	err := db.c.QueryRow(query, userid).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil // Nessun utente trovato
		}
		return false, "", err // C'è stato un errore
	}

	// Se l'utente è stato trovato, restituiamo true e l'username
	return true, username, nil
}
