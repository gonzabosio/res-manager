package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
	resource := new(model.Resource)
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid resource data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(resource); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	creatorIdStr := chi.URLParam(r, "user-id")
	fmt.Println("Creator ID:", creatorIdStr)
	creatorId, err := strconv.ParseInt(creatorIdStr, 10, 64)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to parse user id",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	resource.LockedBy = creatorId
	err = h.Service.CreateResource(resource)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Failed resource creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":  "Resource created successfully",
		"resource": resource,
	}, http.StatusOK)
}

func (h *Handler) GetResourcesBySectionID(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "section-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	resources, err := h.Service.ReadResourcesBySectionID(int64(id))
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed reading resources",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if len(*resources) == 0 {
		WriteJSON(w, map[string]string{
			"message": "No resources found ",
		}, http.StatusOK)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":   "Resources retrieved successfully",
		"resources": resources,
	}, http.StatusOK)
}

func (h *Handler) ModifyResource(w http.ResponseWriter, r *http.Request) {
	resource := new(model.PatchResource)
	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid resource data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err = validate.Struct(resource); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateResource(resource)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not update resource",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":  "Resource updated successfully",
		"resource": resource,
	}, http.StatusOK)
}

func (h *Handler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "resource-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteResourceByID(int64(id))
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not delete resource",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message": "Resource deleted successfully",
	}, http.StatusOK)
}

func (h *Handler) LockResource(w http.ResponseWriter, r *http.Request) {
	reqBody := new(model.LockResourceReq)
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to decode request body",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(reqBody); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	lockStatus, err := h.Service.CheckAndLockResource(reqBody.UserId, reqBody.ResourceId)
	if err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to check or lock resource",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message":     fmt.Sprintf("Resource %d locked by %d", reqBody.ResourceId, reqBody.UserId),
		"lock_status": lockStatus,
	}, http.StatusOK)
}

func (h *Handler) UnlockResource(w http.ResponseWriter, r *http.Request) {
	reqBody := new(model.LockResourceReq)
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to decode request body",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(reqBody); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := h.Service.UnlockResource(reqBody.ResourceId); err != nil {
		WriteJSON(w, map[string]string{
			"message": "Failed to unlock resource",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, map[string]string{
		"message": "resource unlocked successfully",
	}, http.StatusOK)
}
