package database

import (
	"database/sql"
	"errors"
)

func (db *appdbimpl) GetIdUser(name string) (int, error) {
	var user_id UserIdDatabase
	err := db.c.QueryRow("SELECT user_id FROM users WHERE name = ?", name).Scan(&user_id.User_ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			DefaultImage := "https://www.strasys.uk/wp-content/uploads/2022/02/Depositphotos_484354208_S.jpg"
			result, err := db.c.Exec("INSERT INTO users (name, profile_image) VALUES (?, ?)", name, DefaultImage)
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
