package database

// ------------ #OTTENGO LA LISTA DEGLI UTENTI CON USERNAME CERCATO# --------------
func (db *appdbimpl) GetListUsers(name string) ([]Users, error) {
	stringName := "%" + name + "%"

	rows, err := db.c.Query("SELECT user_id, name, profile_image FROM users WHERE name LIKE ?", stringName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []Users
	for rows.Next() {
		var user Users
		err := rows.Scan(&user.UserID, &user.Name, &user.ProfileImage)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ------------ #OTTENGO LA LISTA DEGLI UTENTI DI UNA CONVERSAZIONE# --------------
func (db *appdbimpl) GetConversationUsers(conversationID int) ([]Users, error) {
	rows, err := db.c.Query(
		`SELECT u.user_id, u.name, u.profile_image 
         FROM users u 
         JOIN conversation_participants cm ON u.user_id = cm.user_id 
         WHERE cm.conversation_id = ?`,
		conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []Users
	for rows.Next() {
		var user Users
		err := rows.Scan(&user.UserID, &user.Name, &user.ProfileImage)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
