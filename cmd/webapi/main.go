/*
Webapi is the executable for the main web server.
It builds a web server around APIs from `service/api`.
Webapi connects to external resources needed (database) and starts two web servers: the API web server, and the debug.
Everything is served via the API web server, except debug variables (/debug/vars) and profiler infos (pprof).

Usage:

	webapi [flags]

Flags and configurations are handled automatically by the code in `load-configuration.go`.

Return values (exit codes):

	0
		The program ended successfully (no errors, stopped by signal)

	> 0
		The program ended due to an error

Note that this program will update the schema of the database to the latest version available (embedded in the
executable during the build).
*/
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oooJordan/WasaText/service/api"
	"github.com/oooJordan/WasaText/service/database"
	"github.com/oooJordan/WasaText/service/globaltime"
	"github.com/sirupsen/logrus"
)

// main is the program entry point. The only purpose of this function is to call run() and set the exit code if there is
// any error
func main() {
	//chiama run() per eseguire la funzione
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

// run executes the program. The body of this function should perform the following steps:
// * reads the configuration
// * creates and configure the logger
// * connects to any external resources (like databases, authenticators, etc.)
// * creates an instance of the service/api package
// * starts the principal web server (using the service/api.Router.Handler() for HTTP handlers)
// * waits for any termination event: SIGTERM signal (UNIX), non-recoverable server error, etc.
// * closes the principal web server
func run() error {
	rand.Seed(globaltime.Now().UnixNano())
	// carica la configurazione
	cfg, err := loadConfiguration()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return err
	}

	// inizializza un Logger, per la gestione dei log dell'applicazione
	/*
		COSA È UN LOGGER?
			un logger è uno strumento utile per tracciare e registrare informazioni
			(come messaggi di debug, errori, avvisi, ecc.) che possono essere utili per
			il monitoraggio o la diagnosi durante lo sviluppo e l'esecuzione dell'applicazione.
	*/
	logger := logrus.New()
	logger.SetOutput(os.Stdout) //imposta output logger (stdout)
	if cfg.Debug {              //se la modalità di debug è attiva nella configurazione
		logger.SetLevel(logrus.DebugLevel) // Imposta il livello di log a DebugLevel (+ dettagliato)
	} else {
		logger.SetLevel(logrus.InfoLevel) // Altrimenti, imposta il livello di log a InfoLevel
	}
	// Registra un messaggio di info che indica che l'applicazione sta iniziando
	logger.Infof("application initializing")

	// Connessione al database utilizzando percorso definito nella configurazione (cfg.DB.Filename)
	logger.Println("initializing database support")
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename) // Apre una connessione al database SQLite
	if err != nil {                                     //se ci sono errori termina
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping") // Log un messaggio di debug quando la connessione al database sta per essere chiusa
		_ = dbconn.Close()                //chiude connessione
	}()
	db, err := database.New(dbconn) // Crea una nuova istanza del database usando la connessione appena aperta
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return fmt.Errorf("creating AppDatabase: %w", err)
	}

	// Start (main) API server
	logger.Info("initializing API server")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// viene creato un API router, configurato con i log e la connessione al database,
	// permettendo al router di accedere ai dati e di loggare le informazioni sulle richieste
	// gestisce le rotte (navigatore)
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: db, //connessione al db
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("creating the API server instance: %w", err)
	}
	router := apirouter.Handler() //gestisce richieste HTTP (manipolatore)

	//per controllare se è necessario aggiungere rotte extra per un Web UI
	router, err = registerWebUI(router)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return fmt.Errorf("registering web UI handler: %w", err)
	}

	// Apply CORS policy
	// il router applicherà le politiche sulle richieste che riceve
	router = applyCORSHandler(router)

	// Creazione di API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,      //indirizzo in cui il server ascolta
		Handler:           router,               //router che gestisce le richieste
		ReadTimeout:       cfg.Web.ReadTimeout,  //timeout per la lettura di una richiesta
		ReadHeaderTimeout: cfg.Web.ReadTimeout,  //timaout per la scrittura di una richiesta
		WriteTimeout:      cfg.Web.WriteTimeout, //timeout per inviare una risposta
	}

	// Start the service listening for requests in a separate goroutine
	go func() {
		logger.Infof("API listening on %s", apiserver.Addr)
		// Avvia il server API. Questa funzione è bloccante, quindi invia
		// l'errore nel canale `serverErrors` se qualcosa va storto
		serverErrors <- apiserver.ListenAndServe()
		logger.Infof("stopping API server")
	}()

	// Waiting for shutdown signal or POSIX signals
	select {
	case err := <-serverErrors: //errore non recuperabile nel server
		// Non-recoverable server error
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown: //segnale di terminazione
		logger.Infof("signal %v received, start shutdown", sig)

		// chiude il router, segnala che non deve più ricevere nuove richieste
		err := apirouter.Close()
		if err != nil {
			logger.WithError(err).Warning("graceful shutdown of apirouter error")
		}

		// viene creato un contesto con un timeout
		// limite di tempo entro cui completare le richieste in sospeso
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// chiede al server di terminare la gestione delle richieste
		err = apiserver.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Warning("error during graceful shutdown of HTTP server")
			err = apiserver.Close()
		}

		// gestione dello stato di arresto del server
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}
	//terminazione del server senza errori
	return nil
}
