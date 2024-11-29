package api

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var validNickname = regexp.MustCompile("^[a-zA-Z0-9_]+$")

//var emojiCodeRegex = regexp.MustCompile(`^[A-Za-z0-9+]+$`)

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
func (rt *_router) IsValidToken(r *http.Request, w http.ResponseWriter) (bool, error) {
	//valore dell'intestazione Authorization della richiesta HTTP
	authHeader := r.Header.Get("Authorization")
	fmt.Println(authHeader + "\n")

	//controllo se non è vuoto
	if len(authHeader) == 0 {
		http.Error(w, "Authorization Header missing", http.StatusForbidden)
		return false, errors.New("authorization Header missing")
	}
	//divido la stringa in base agli spazi es:["Barer", "user"] e cnontrollo se c'è barer
	HeaderPart := strings.Fields(authHeader)
	if len(HeaderPart) < 2 || HeaderPart[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusForbidden)
		return false, errors.New("invalid Authorization format")
	}
	token := HeaderPart[1] //prendo il token

	userid := extractUserIdFromToken(token) //id_utente

	found, err := rt.db.CheckIDDatabase(userid)
	if err != nil || !found {
		http.Error(w, "Non autorizzato", http.StatusUnauthorized)
		return false, nil
	}
	return true, nil
}

func extractUserIdFromToken(token string) int {
	id, _ := strconv.Atoi(token)
	return id
}
