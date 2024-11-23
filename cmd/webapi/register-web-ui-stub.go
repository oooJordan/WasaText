//go:build !webui

package main

import (
	"net/http"
)

// Questa funzione non fa nulla, restituisce solo il router come è.
// perchè 'webui' non è stata specificata
// serve a aggiungere rotte (URL) specifiche per l'interfaccia utente web se è necessaria
func registerWebUI(hdl http.Handler) (http.Handler, error) {
	return hdl, nil
}
