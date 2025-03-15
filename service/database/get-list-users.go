package database

// ------------ #OTTENGO LA LISTA DEGLI UTENTI CON USERNAME CERCATO# --------------
func (db *appdbimpl) GetListUsers(name string) ([]Users, error) {

	stringName := "%" + name + "%"
	// Primo passo: seleziono i nomi corrispondenti
	rows, err := db.c.Query("SELECT name FROM Users WHERE name LIKE ?", stringName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var userName string
		if err := rows.Scan(&userName); err != nil {
			return nil, err
		}
		names = append(names, userName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Secondo passo: per ogni nome, recupero le informazioni complete
	var users []Users
	for _, userName := range names {
		var user Users
		err := db.c.QueryRow("SELECT name, user_id, profile_image FROM Users WHERE name = ?", userName).Scan(&user.Name, &user.UserID, &user.ProfileImage)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
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
	return users, nil
}
