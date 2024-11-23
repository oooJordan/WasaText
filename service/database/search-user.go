package database

func (db *appdbimpl) SearchUser(queryStr string) ([]User, error) {
	/*
				query SQL per trovare i record nella tabella User dove il campo nickname
		    	contiene la stringa passata come parametro
	*/
	nickname, err := db.c.Query("SELECT * FROM User WHERE nickname LIKE ?", "%"+queryStr+"%")
	if err != nil {
		nickname.Close()
		return nil, err
	}
	// Controlla se ci sono errori nel risultato della query
	if nickname.Err() != nil {
		return nil, nickname.Err()
	}

	var users []User      //array di utenti
	for nickname.Next() { //itero su risultati della query
		var u User
		//mappa risultato query su capi della struttura User
		err = nickname.Scan(&u.User_ID, &u.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, u) //appende l'utente nella lista
	}
	// Restituisce lista utenti trovati e nessun errore
	return users, nil

}
