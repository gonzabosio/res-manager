package handlers

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
