package http

import (
	"encoding/json"
	"net/http"

	"github.com/zkhrg/go_day03/internal/api"
)

func generateTokenHandler(a *api.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := a.GetTokenByName(r.Context().Value(UsernameContextKey).(string))
		if err != nil {
			http.Error(w, "Failed to get username", http.StatusInternalServerError)
			return
		}

		token, err := a.GetTokenByName(username)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
