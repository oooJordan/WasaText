package database

import (
	"database/sql"
	"errors"
)

var ErrUsernameAlreadyInUse = errors.New("username already in use")

// ----------- #AGGIORNA IMMAGINE PROFILO UTENTE# ------------------
func (db *appdbimpl) UpdateProfileImage(userid int, NewImage string) error {
	_, err := db.c.Exec("UPDATE users SET profile_image = ? WHERE user_id = ?", NewImage, userid)
	if err != nil {
		return err
	}

	return nil
}

// ----------- #OTTIENE IMMAGINE PROFILO UTENTE# ------------------
func (db *appdbimpl) GetProfileImage(userid int) (string, error) {
	var imageURL string
	err := db.c.QueryRow("SELECT profile_image FROM users WHERE user_id = ?", userid).Scan(&imageURL)
	if err != nil {
		return "", err
	}
	return imageURL, nil
}

func (db *appdbimpl) GetIdUser(name string, defaultimage string) (int, error) {
	var user_id UserIdDatabase
	err := db.c.QueryRow("SELECT user_id FROM users WHERE name = ?", name).Scan(&user_id.User_ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result, err := db.c.Exec("INSERT INTO users (name, profile_image) VALUES (?, ?)", name, defaultimage)
			if err != nil {
				return 0, err
			}
			newId, err := result.LastInsertId()
			if err != nil {
				return 0, err
			}
			user_id.User_ID = int(newId)
		}
	}
	return user_id.User_ID, nil
}

// ------------------ #AGGIORNA USERNAME UTENTE# --------------------------
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
