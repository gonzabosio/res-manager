package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	section := new(model.Section)
	if err := json.NewDecoder(r.Body).Decode(&section); err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid section data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(section); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	id, err := h.Service.CreateSection(section)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Failed section creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":    "Section created successfully",
		"section_id": id,
	}, http.StatusOK)
}

func (h *Handler) GetSectionsByProjectID(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "project-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	sections, err := h.Service.ReadSectionsByProjectID(int64(id))
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed reading sections",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if len(*sections) == 0 {
		WriteJSON(w, map[string]string{
			"message": "No sections found",
		}, http.StatusOK)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":  "Sections retrieved successfully",
		"sections": sections,
	}, http.StatusOK)
}

func (h *Handler) ModifySection(w http.ResponseWriter, r *http.Request) {
	section := new(model.PutSection)
	err := json.NewDecoder(r.Body).Decode(&section)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid section data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateSection(section)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not update section",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message": "Section updated successfully",
		"section": section,
	}, http.StatusOK)
}

func (h *Handler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "section-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteSectionByID(int64(id))
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not delete section",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message": "Section deleted successfully",
	}, http.StatusOK)
}
