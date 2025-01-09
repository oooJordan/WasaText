package database

func (db *appdbimpl) UpdateProfileImage(userid int, NewImage string) error {
	_, err := db.c.Exec("UPDATE users SET profile_image = ? WHERE user_id = ?", NewImage, userid)
	if err != nil {
		return err
	}

	return nil
}
