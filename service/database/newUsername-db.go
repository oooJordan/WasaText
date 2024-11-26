package database

import "errors"

func (db *appdbimpl) UpdateUsername(userid int, newUsername string) error {
	// Controlla se il nuovo username esiste già
	var existing int
	err := db.c.QueryRow("SELECT COUNT(*) FROM Users WHERE name = ?", newUsername).Scan(&existing)
	if err != nil {
		return err
	}
	if existing > 0 {
		return errors.New("username already in use")
	}

	// Aggiorna il nome utente
	_, err = db.c.Exec("UPDATE Users SET name = ? WHERE user_id = ?", newUsername, userid)
	if err != nil {
		return err
	}

	return nil
}
