package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/config"
	"github.com/gonzabosio/res-manager/model"
)

func (h *Handler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	googleConfig := config.GoogleConfig()
	url := googleConfig.AuthCodeURL("randomstate")
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *Handler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state != "randomstate" {
		http.Error(w, "states don't match", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	googleConfig := config.GoogleConfig()
	log.Println(code)
	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Code-Token Exchange Failed", http.StatusInternalServerError)
		return
	}
	// fetch user data with access token
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "User Data Fetch Failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	log.Printf("RefreshToken: %v - AccessToken: %v\n", token.RefreshToken, token.AccessToken)
	// log.Printf("ExpiresIn: %v - Expiry: %v\n", token.ExpiresIn, token.Expiry)
	http.SetCookie(w, &http.Cookie{
		Name:    "access-token",
		Value:   token.AccessToken,
		Path:    "/",
		Expires: token.Expiry,
	})
	refreshURL := fmt.Sprintf(`<html><head><meta http-equiv="refresh" content="0;url=%v/refresh"></head><body></body></html>`, os.Getenv("FRONT_URL"))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, refreshURL)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	user := new(model.User)
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid user data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	inserted, err := h.Service.InsertOrGetUser(user)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Failed user creation",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if inserted {
		WriteJSON(w, map[string]interface{}{
			"message": "User created successfully",
			"user":    user,
		}, http.StatusOK)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message": "User retrieved successfully",
		"user":    user,
	}, http.StatusOK)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	user := new(model.User)
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteJSON(w, map[string]string{
			"message": "Invalid user data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	// err := h.Service.VerifyUser(user)
	// if err != nil {
	// 	WriteJSON(w, map[string]string{
	// 		"message": "Failed verifying user",
	// 		"error":   err.Error(),
	// 	}, http.StatusBadRequest)
	// 	return
	// }
	WriteJSON(w, map[string]interface{}{
		"message": "User logged in successfully",
		"user":    user,
	}, http.StatusOK)
}

// func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
// 	users, err := h.Service.ReadUsers()
// 	if err != nil {
// 		WriteJSON(w, map[string]string{
// 			"message": "Failed reading users",
// 			"error":   err.Error(),
// 		}, http.StatusBadRequest)
// 		return
// 	}
// 	if len(*users) == 0 {
// 		WriteJSON(w, map[string]string{
// 			"message": "No users found",
// 		}, http.StatusOK)
// 		return
// 	}
// 	WriteJSON(w, map[string]interface{}{
// 		"message": "Users retrieved successfully",
// 		"users":   users,
// 	}, http.StatusOK)
// }

func (h *Handler) ModifyUser(w http.ResponseWriter, r *http.Request) {
	user := new(model.PatchUser)
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Invalid user data or bad format",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	if err := validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		WriteJSON(w, map[string]string{
			"message": "Validation error",
			"error":   errors.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.UpdateUser(user)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not update user",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message": "User updated successfully",
		"user":    user,
	}, http.StatusOK)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idS := chi.URLParam(r, "user-id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not convert id",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteUserByID(int64(id))
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"message": "Could not delete user",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, map[string]interface{}{
		"message": "User deleted successfully",
	}, http.StatusOK)
}
