package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/broscoe7/go-school-api/internal/models"
	"github.com/broscoe7/go-school-api/internal/repository/sqlconnect"
	"github.com/broscoe7/go-school-api/pkg/utils"
)

func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	var newExecs []models.Exec
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&newExecs)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, t := range newExecs {
		err := CheckBlankFields(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedExecs, err := sqlconnect.AddExecsDBHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)
}

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	var execs []models.Exec
	execs, err := sqlconnect.GetExecsDBHandler(execs, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(execs),
		Data:   execs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetExecHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println(err)
		return
	}

	exec, err := sqlconnect.GetExecByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]any
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PatchExecs(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func PatchExecHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedExec, err := sqlconnect.PatchExecDB(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedExec)
}

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteExecDB(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Exec

	// data validation
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// check that user exists
	user, err := sqlconnect.GetExecByUsername(req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}

	// check if user is active
	if user.InactiveStatus {
		http.Error(w, "Account is not active", http.StatusForbidden)
		return
	}

	// verify password
	err = utils.VerifyPassword(req.Password, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	// generate token
	jwt, err := utils.SignToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Could not create login token", http.StatusInternalServerError)
		return
	}

	// send token as response or cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    jwt,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	}{
		Token: jwt,
	}
	json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Logged out successfully"}`))
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid exec id", http.StatusBadRequest)
		return
	}

	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := sqlconnect.UpdatePasswordInDB(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// send token as response or cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}{
		Message: "Password successfully updated",
		Token:   token,
	}
	json.NewEncoder(w).Encode(response)
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = sqlconnect.ForgotPasswordDBHandler(req.Email)
	if err != nil {
		http.Error(w, "could not send password reset email", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "A password reset link has been sent to %s", req.Email)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	// validation
	token := r.PathValue("resetcode")
	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.NewPassword == "" || req.ConfirmPassword == "" || token == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		http.Error(w, "Passwords must match", http.StatusBadRequest)
	}

	// reset password
	err = sqlconnect.ResetPasswordDBHandler(token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "Password successfully reset")
}
