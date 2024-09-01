package api

import (
	"log"

	"github.com/zkhrg/go_day03/internal/auth"
)

func (a *API) GetTokenByName(username string) (string, error) {
	token, err := auth.GetTokenByName(username)
	if err != nil {
		log.Printf("error with generating token")
		return "", err
	}
	return token, nil
}

func (a *API) ValidateToken(token string) bool {
	if _, err := auth.ValidateToken(token); err != nil {
		return false
	}
	return true
}
