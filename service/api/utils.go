package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

var validNickname = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func IsValidNickname(nick string) bool {
	if len(nick) < 3 || len(nick) > 16 {
		return false
	}

	return validNickname.MatchString(nick)
}

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

// ------------#CARICAMENTO IMMAGINE E CREAZIONE URL#----------------
const uploadDir = "/home/jordan/Documents/university/WASA/WasaText/foto"

func (rt *_router) uploadImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	isValid, userID, err := rt.IsValidToken(r, w)
	if err != nil || !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10MB di limite
	if err != nil {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Controllo il formato dell'immagine (PNG, JPEG, JPG)
	allowedFormats := map[string]bool{"image/png": true, "image/jpeg": true, "image/jpg": true}
	fileHeader := make([]byte, 512)
	_, err = file.Read(fileHeader)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(fileHeader)
	if !allowedFormats[contentType] {
		http.Error(w, "Unsupported file format", http.StatusUnsupportedMediaType)
		return
	}

	// Resetto il puntatore del file per poterlo leggere di nuovo
	file.Seek(0, io.SeekStart)

	// Genero un nome univoco per il file
	fileExt := filepath.Ext(handler.Filename)
	if fileExt == "" {
		fileExt = ".png" // Default a PNG se l'estensione non è riconosciuta
	}
	timestamp := time.Now().Format("20060102150405")
	fileName := strconv.Itoa(userID) + "_" + timestamp + fileExt

	// Creazione della directory dell'utente (se non esiste)
	userDir := filepath.Join(uploadDir, strconv.Itoa(userID))
	err = os.MkdirAll(userDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Salvo l'immagine
	filePath := filepath.Join(userDir, fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	// Genero l'URL per visualizzare l'immagine
	host := r.Host
	scheme := "https" // Default -> HTTPS

	if r.TLS == nil { // Se non è HTTPS -> HTTP
		scheme = "http"
	}
	// URL dell'immagine -> /foto/{userID}/{fileName}
	imageURL := scheme + "://" + host + "/foto/" + strconv.Itoa(userID) + "/" + fileName

	// Risposta JSON con l'URL dell'immagine
	response := map[string]string{
		"imageUrl": imageURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ------------#CONTROLLO SE L'UTENTE È NELLA CONVERSAZIONE#----------------
func (rt *_router) isUserInConversation(conversationID int, userID int, chatType string) (bool, error) {
	switch chatType {
	case "group_chat":
		return rt.db.IsUserInGroup(conversationID, userID)
	case "private_chat":
		return rt.db.IsUserInPrivateChat(conversationID, userID)
	default:
		return false, errors.New("invalid conversation type")
	}
}

func getIntParam(w http.ResponseWriter, ps httprouter.Params, paramName string) (int, bool) {
	paramStr := ps.ByName(paramName)
	param, err := strconv.Atoi(paramStr)
	if err != nil {
		http.Error(w, "Invalid "+paramName, http.StatusBadRequest)
		return 0, false
	}
	return param, true
}
