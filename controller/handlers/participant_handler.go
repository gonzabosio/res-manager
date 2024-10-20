package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	participant := new(model.Participant)
	if err := json.NewDecoder(r.Body).Decode(&participant); err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid participant data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	id, err := h.Service.InsertParticipant(participant)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Failed participant creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":   "Participant added successfully",
		"object_id": id,
	}, http.StatusOK)
}

func (h *Handler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "team-id")
	teamId, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	participants, err := h.Service.ReadParticipants(int64(teamId))
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed reading participants",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":      "Participants retrieved successfully",
		"participants": participants,
	}, http.StatusOK)
}

func (h *Handler) DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteParticipantByID(int64(userId))
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Could not delete participant",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]string{
		"message": "Participant deleted successfully",
	}, http.StatusOK)
}
