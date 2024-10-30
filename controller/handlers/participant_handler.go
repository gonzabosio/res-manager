package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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
	if err := validate.Struct(participant); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	inserted, err := h.Service.RegisterParticipant(participant)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Failed to create or return participant data",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if inserted {
		WriteJSON(w, map[string]interface{}{
			"message":     "Participant added successfully",
			"participant": participant,
		}, http.StatusOK)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":     "Participant retrieved successfully",
		"participant": participant,
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
	idS := chi.URLParam(r, "user-id")
	userId, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert user id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	idS = chi.URLParam(r, "team-id")
	teamId, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert team id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteParticipantByIDs(int64(userId), int64(teamId))
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
