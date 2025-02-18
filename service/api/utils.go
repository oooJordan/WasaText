package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

var validNickname = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// var emojiCodeRegex = regexp.MustCompile(`^[A-Za-z0-9+]+$`)

func IsValidNickname(nick string) bool {
	if len(nick) < 3 || len(nick) > 16 {
		return false
	}

	return validNickname.MatchString(nick)
}

/*
	func isValidMessage(mex string) bool {
		if len(mex) < 1 || len(mex) > 200 {
			return false
		}
		return true
	}

	func isValidEmoji(emojicode string) bool {
		if len(emojicode) < 1 || len(emojicode) > 20 {
			return false
		}
		return emojiCodeRegex.MatchString(emojicode)
	}
*/
func (rt *_router) IsValidToken(r *http.Request, w http.ResponseWriter) (bool, int, error) {
	// valore dell'intestazione Authorization della richiesta HTTP
	authHeader := r.Header.Get("Authorization")

	// controllo se non è vuoto
	if len(authHeader) == 0 {
		http.Error(w, "Authorization Header missing", http.StatusForbidden)
		return false, 0, errors.New("authorization Header missing")
	}
	// divido la stringa in base agli spazi es:["Barer", "user"] e cnontrollo se c'è barer
	HeaderPart := strings.Fields(authHeader)
	if len(HeaderPart) < 2 || HeaderPart[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusForbidden)
		return false, 0, errors.New("invalid Authorization format")
	}
	token := HeaderPart[1] // prendo il token

	userid := ExtractUserIdFromToken(token) // id_utente

	found, _, err := rt.db.CheckIDDatabase(userid)
	if err != nil || !found {
		http.Error(w, "Non autorizzato", http.StatusUnauthorized)
		return false, 0, nil
	}
	return true, userid, nil
}

func ExtractUserIdFromToken(token string) int {
	id, _ := strconv.Atoi(token)
	return id
}

func (rt *_router) uploadImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	err := r.ParseMultipartForm(10 << 20) // Max 10MB
	if err != nil {
		ctx.Logger.WithError(err).Error("uploadFile: error parsing multipart form")
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	// estraggo il file dalla richiesta http
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		ctx.Logger.WithError(err).Error("uploadFile: error retrieving file from form")
		http.Error(w, "File not found in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// directory di destinazione
	destDir := "/home/jordan/Documents/university/WASA/WasaText/images"
	// se non esiste viene creata
	if err := os.MkdirAll(destDir, 0755); err != nil {
		ctx.Logger.WithError(err).Error("uploadFile: error creating destination directory")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// uso il nome originale del file per creare il percorso finale
	uniqueFileName := fileHeader.Filename
	destPath := filepath.Join(destDir, uniqueFileName)
	// creo file nella directory di destinazione
	dst, err := os.Create(destPath)
	if err != nil {
		ctx.Logger.WithError(err).Error("uploadFile: error creating destination file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	// copio il contenuto del file ricevuto nel file di destinazione
	if _, err := io.Copy(dst, file); err != nil {
		ctx.Logger.WithError(err).Error("uploadFile: error saving file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// url pubblico
	publicURL := "http://localhost:3000/foto/" + uniqueFileName

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	// url pubblico in formato 	json
	json.NewEncoder(w).Encode(map[string]string{
		"imageUrl": publicURL,
	})
}

// directory con foto
const BaseDirFoto = "/home/jordan/Documents/university/WASA/WasaText/foto"

// fileExists verifica se il file esiste nel percorso specificato
func fileExists(path string) bool {
	_, err := os.Stat(path) // controlla lo stato del file
	return err == nil       // true se il file esiste
}

// verifica se l'URL è valido, appartiene alla cartella "/foto/" e se il file esiste localmente
func validateLocalImageURL(imageURL string) bool {
	u, err := url.Parse(imageURL) // analizza la stringa per ottenere url
	// controlla se l'url è valido
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	// controllo se il percorso inizi con "/foto/"
	if !strings.HasPrefix(u.Path, "/foto/") {
		return false
	}

	// estrae il nome del file rimuovendo "/foto/"
	fileName := strings.TrimPrefix(u.Path, "/foto/")
	if fileName == "" {
		return false
	}

	// costruisce il percorso locale del file
	localPath := filepath.Join(BaseDirFoto, fileName)
	// verifica se il file esiste
	return fileExists(localPath)
}
