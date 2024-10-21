package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := new(model.User)
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Invalid user data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	hashedPw, err := hashPassword([]byte(user.Password))
	if err != nil {
		writeJSON(w, map[string]string{
			"message": "Could not hash password",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	user.Password = hashedPw
	id, err := h.Service.InsertUser(user)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Failed user creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "User created successfully",
		"user_id": id,
	}, http.StatusOK)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	user := new(model.User)
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, map[string]string{
			"message": "Invalid user data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	err := h.Service.VerifyUser(user)
	if err != nil {
		writeJSON(w, map[string]string{
			"message": "Failed verifying user",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "User logged in successfully",
		"user":    user,
	}, http.StatusOK)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.ReadUsers()
	if err != nil {
		writeJSON(w, map[string]string{
			"message": "Failed reading users",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if len(*users) == 0 {
		writeJSON(w, map[string]string{
			"message": "No users found",
		}, http.StatusOK)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "Users retrieved successfully",
		"users":   users,
	}, http.StatusOK)
}

func (h *Handler) ModifyUser(w http.ResponseWriter, r *http.Request) {
	user := new(model.User)
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Invalid user data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		writeJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateUser(user)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not update user",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "User updated successfully",
		"user":    user,
	}, http.StatusOK)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "user-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteUserByID(int64(id))
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"message": "Could not delete user",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"message": "User deleted successfully",
	}, http.StatusOK)
}
