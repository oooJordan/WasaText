/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"log"
)

// in AppDatabase ci sono i metodi definiti in database.go
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error
	GetIdUser(name string) (int, error)
	GetListUsers(name string) ([]Users, error)
	UpdateUsername(userid int, NewUsername string) error
	CheckIDDatabase(userid int) (bool, error)
	//CheckUser(User) (User, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) { //inizializza il database
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// attiva le foreign keys
	_, errPramga := db.Exec(`PRAGMA foreign_keys= ON`)
	if errPramga != nil {
		return nil, errPramga
	}

	// controlla se il database esiste. se non esiste viene creata la struttura
	var tableName string //variabile per memorizzare nome tabella
	// viene eseguita una query SQL sul database per verificare se esiste una tabella chiamata example_table (ritorna 1 riga)
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) { //se non ha trovato nessuna tabella la crea

		// Creazione DB per gli Users se non esiste
		users := `CREATE TABLE users 
						(user_id INTEGER NOT NULL, 
						name VARCHAR(25) NOT NULL UNIQUE,
						profile_image TEXT,
						PRIMARY KEY("user_id" AUTOINCREMENT));`

		conversations := `CREATE TABLE conversations (
						ConversationId INTEGER NOT NULL PRIMARY KEY, 
						chatType TEXT NOT NULL, 
						groupName TEXT, 
						imageGroup TEXT, 
						authorId INTEGER NOT NULL, 
						timestamp DATETIME NOT NULL, 
						FOREIGN KEY(authorId) REFERENCES users(user_id) );`

		participants := `CREATE TABLE conversation_participants (
							ConvId INTEGER NOT NULL, 
							user_id INTEGER NOT NULL, 
							UNIQUE(ConvId, user_id), 
							FOREIGN KEY(ConvId) REFERENCES conversations(ConversationId),
							FOREIGN KEY(user_id) REFERENCES users(user_id)
						);`

		messages := `CREATE TABLE messages (
								id INTEGER NOT NULL PRIMARY KEY, 
								ConversationId INTEGER NOT NULL, 
								sender_id INTEGER NOT NULL, 
								content TEXT NOT NULL, 
								timestamp DATETIME NOT NULL, 
								is_read BOOLEAN DEFAULT FALSE, 
								FOREIGN KEY(ConversationId) REFERENCES conversations(ConvId),
								FOREIGN KEY(sender_id) REFERENCES users(user_id)
							);`

		reactions := `CREATE TABLE message_reactions (
								message_id INTEGER NOT NULL, 
								user_id INTEGER NOT NULL, 
								reaction TEXT NOT NULL, 
								UNIQUE(message_id, user_id), 
								FOREIGN KEY(message_id) REFERENCES messages(ConvId),
								FOREIGN KEY(user_id) REFERENCES users(user_id)
							);`

		_, err = db.Exec(users)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(conversations)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(participants)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(messages)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(reactions)
		if err != nil {
			log.Fatal(err)
		}

		/*
			Il valore di ritorno non viene utilizzato (_), perché non ci interessa
			il risultato della creazione della tabella, ma solo sapere se ci sono
			errori nell'esecuzione
		*/

	}
	// Crea un nuovo appdbimpl con la connessione al database
	return &appdbimpl{
		c: db, // Inizializza il campo c con la connessione al database
	}, nil
}

// verifica se la connessione al database è attiva e funzionante inviando un ping
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
	// db.c rappresenta la connessione al database
}
