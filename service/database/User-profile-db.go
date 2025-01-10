package database

func (db *appdbimpl) UpdateProfileImage(userid int, NewImage string) error {
	_, err := db.c.Exec("UPDATE users SET profile_image = ? WHERE user_id = ?", NewImage, userid)
	if err != nil {
		return err
	}

	return nil
}

func (db *appdbimpl) GetProfileImage(userid int) (string, error) {
	var imageURL string
	err := db.c.QueryRow("SELECT profile_image FROM users WHERE user_id = ?", userid).Scan(&imageURL)
	if err != nil {
		return "", err
	}
	return imageURL, nil
}
