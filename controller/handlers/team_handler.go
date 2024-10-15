package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	team := new(model.Team)
	json.NewDecoder(r.Body).Decode(&team)
	id, err := h.RepositoryService.CreateTeam(team)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]string{
		"message": "Team created successfully",
		"team_id": id,
	}, http.StatusOK)
}
