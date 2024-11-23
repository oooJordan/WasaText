package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

// WebAPIConfiguration describes the web API configuration. This structure is automatically parsed by
// loadConfiguration and values from flags, environment variable or configuration file will be loaded.
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:/conf/config.yml"` //posizione del file YAML di configurazione
	}
	Web struct { //configurazione webserver
		APIHost   string `conf:"default:0.0.0.0:3000"` //porta in cui API è in ascolto
		DebugHost string `conf:"default:0.0.0.0:4000"` //indirizzo e porta per l'host di debug
		//tempi di timeout per le operazioni di lettura, scrittura e spegnimento
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Debug bool //flag per attivare o disattivare mod debug
	DB    struct {
		Filename string `conf:"default:/tmp/decaf.db"`
	}
}

// loadConfiguration creates a WebAPIConfiguration starting from flags, environment variables and configuration file.
// It works by loading environment variables first, then update the config using command line flags, finally loading the
// configuration file (specified in WebAPIConfiguration.Config.Path).
// So, CLI parameters will override the environment, and configuration file will override everything.
// Note that the configuration file can be specified only via CLI or environment variable.
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// La libreria conf si occupa di parsare gli argomenti passati alla riga di
	// comando e di popolare la struttura WebAPIConfiguration
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("generating config usage: %w", err)
			}
			fmt.Println(usage) //nolint:forbidigo
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	/*
		Dopo aver gestito i parametri dalla riga di comando e le variabili d'ambiente,
		se la configurazione prevede un file YAML, il programma tenta di aprirlo e caricarlo.
		Se il file esiste, viene letto e deserializzato in una struttura Go
	*/
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("can't read the config file, while it exists: %w", err)
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp) //legge file YAML
		if err != nil {                 // se è diverso c'è errore
			return cfg, fmt.Errorf("can't read config file: %w", err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg) //file deserializzato in un oggetto Go
		if err != nil {                      // se è diverso c'è errore
			return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
		}
		_ = fp.Close()
	}
	/* Se tutto va a buon fine, la funzione restituisce la configurazione
	(WebAPIConfiguration) e un errore (che in questo caso sarà nil se non
	ci sono problemi). Se c'è un errore, la funzione lo restituisce insieme
	a un messaggio descrittivo */
	return cfg, nil
}
