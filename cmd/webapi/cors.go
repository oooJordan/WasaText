package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		/*questa parte permette che le richieste includano le due intestazioni
		Utile per la gestione dei token (come nel caso di JWT) e specificare il
		tipo di contenuto delle richieste (es application/json)*/
		handlers.AllowedHeaders([]string{
			"Authorization",
			"content-type",
		}),
		//metodi permessi
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Do not modify the CORS origin and max age, they are used in the evaluation.
		handlers.AllowedOrigins([]string{"*"}), //permette richieste provenienti da qualsiasi dominio
		handlers.MaxAge(1),                     //quanto tempo le informazioni CORS sono memorizzate nella cache del browser
	)(h)
}

/*
 Se non configurassi CORS, il browser rifiuterebbe automaticamente
 le richieste HTTP tra domini differenti.  CORS è una politica di sicurezza
 implementata a livello di browser, ma deve essere gestita e configurata
 correttamente anche dal server. Il server deve specificare quali origini
 possono accedere alle sue risorse
*/
