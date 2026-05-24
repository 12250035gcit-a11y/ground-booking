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
		httpReps.ResponseWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    user.Email,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]interface{}{
		"message": "login success",
		"user":    user,
	})
}

func GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		httpReps.ResponseWithError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	user, err := model.GetUserByEmail(cookie.Value)
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusUnauthorized, "session invalid")
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, user)
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
