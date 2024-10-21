package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	team := new(model.Team)
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Invalid resource data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(team); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	id, err := h.Service.CreateTeam(team)
	if err != nil {
		writeJSON(w, map[string]string{
			"message": "Failed team creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Team created successfully",
		"team_id": id,
	}, http.StatusOK)
}

func (h *Handler) VerifyTeamByName(w http.ResponseWriter, r *http.Request) {
	team := new(model.Team)
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeJSON(w, map[string]string{
			"message": "Invalid team data or bad format",
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(team); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	err := h.Service.ReadTeamByName(team)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not verify the team",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Joined team successfully",
		"team":    team,
	}, http.StatusOK)
}

func (h *Handler) ModifyTeam(w http.ResponseWriter, r *http.Request) {
	team := new(model.Team)
	err := json.NewDecoder(r.Body).Decode(&team)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Invalid team data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateTeam(team)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not update team",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Team updated successfully",
		"team":    team,
	}, http.StatusOK)
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "team-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteTeamByID(int64(id))
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not delete team",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Team deleted successfully",
	}, http.StatusOK)
}
