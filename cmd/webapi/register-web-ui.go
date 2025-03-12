//go:build webui

package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	// "github.com/oooJordan/WasaText/webui"
)

// si occupa di registrare il Web UI (interfaccia utente) nel server web
func registerWebUI(hdl http.Handler) (http.Handler, error) {
	distDirectory, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("error embedding WebUI dist/ directory: %w", err)
	}
	// definito un nuovo handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/dashboard/") {
			http.StripPrefix("/dashboard/", http.FileServer(http.FS(distDirectory))).ServeHTTP(w, r)
			return
		}
		hdl.ServeHTTP(w, r)
	}), nil
}
