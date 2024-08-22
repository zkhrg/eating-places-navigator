package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zkhrg/go_day03/internal/api"
)

func NearestPlacesHandler(a *api.API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lat := r.Context().Value(LatContextKey).(float64)
		lon := r.Context().Value(LonContextKey).(float64)
		response, err := a.Store.GetNearestPlaces(lat, lon)
		if err != nil {
			log.Printf("handler nearest places can not get a nearest places from strore")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
