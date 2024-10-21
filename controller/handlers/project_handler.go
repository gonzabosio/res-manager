package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	proj := new(model.Project)
	if err := json.NewDecoder(r.Body).Decode(&proj); err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Invalid project data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(proj); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	id, err := h.Service.CreateProject(proj)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Failed project creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message":    "Project created successfully",
		"project_id": id,
	}, http.StatusOK)
}

func (h *Handler) GetProjectsByTeamID(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "team-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		writeJSON(w, map[string]string{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	projects, err := h.Service.ReadProjectsByTeamID(int64(id))
	if err != nil {
		writeJSON(w, map[string]string{
			"message": "Failed reading projects",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if len(*projects) == 0 {
		writeJSON(w, map[string]string{
			"message": "No projects found",
		}, http.StatusOK)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Projects retrieved successfully",
		"project": projects,
	}, http.StatusOK)
}
func (h *Handler) ModifyProject(w http.ResponseWriter, r *http.Request) {
	project := new(model.Project)
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Invalid project data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err = validate.Struct(project); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateProject(project)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not update project",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Project updated successfully",
		"team":    project,
	}, http.StatusOK)
}
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "project-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteProjectByID(int64(id))
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not delete project",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Project deleted successfully",
	}, http.StatusOK)
}
