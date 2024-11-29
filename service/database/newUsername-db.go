package database

import "errors"

var ErrUsernameAlreadyInUse = errors.New("username already in use")

func (db *appdbimpl) UpdateUsername(userid int, newUsername string) error {
	// Controllo se il nuovo username esiste già
	var existing int
	err := db.c.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", newUsername).Scan(&existing)
	if err != nil {
		return err
	}
	if existing > 0 {
		return ErrUsernameAlreadyInUse
	}

	// Aggiorno il nome utente
	_, err = db.c.Exec("UPDATE users SET name = ? WHERE user_id = ?", newUsername, userid)
	if err != nil {
		return err
	}

	return nil
}
