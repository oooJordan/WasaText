package api

import (
	"errors"
	"net/http"
	"regexp"
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
func IsValidToken(req *http.Request, w http.ResponseWriter) (string, error) {
	//valore dell'intestazione Authorization della richiesta HTTP
	authHeader := req.Header.Get("Authorization")

	//controllo della presenza dell'intestazione
	if len(authHeader) == 0 {
		http.Error(w, "Authorization Header missing", http.StatusForbidden)
		return "", errors.New("authorization Header missing")
	}
	//divido la stringa in base agli spazi es:["Barer", "user"]
	HeaderPart := strings.Fields(authHeader)
	//se l'array ha meno di 2 elementi vuol dire che manca qualcosa
	if len(HeaderPart) < 2 {
		http.Error(w, "Invalid format Auth", http.StatusForbidden)
		return "", errors.New("invalid format Auth")
	}
	//return ultimo elemento array
	return HeaderPart[len(HeaderPart)-1], nil
}
