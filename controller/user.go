package controller

import (
	"encoding/json"
	"myapp/model"
	"myapp/utils/httpReps"
	"net/http"
)

func Adduser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON Body")
		return
	}
	defer r.Body.Close()

	if err := user.Signup(); err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusCreated, map[string]string{
		"status": "User registered. Awaiting admin approval.",
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON Body")
		return
	}
	defer r.Body.Close()

	err = user.Login()
	if err != nil {
		if err.Error() == "account pending approval" {
			httpReps.ResponseWithError(w, http.StatusForbidden, "account pending approval")
			return
		}
		httpReps.ResponseWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]interface{}{
		"message": "login success",
		"user":    user,
	})
}
