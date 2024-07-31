package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var errEmailRequired = errors.New("email is required")
var errFirstNameRequired = errors.New("first name is required")
var errLastNameRequired = errors.New("last name is required")
var errPasswordRequired = errors.New("password is required")

type UserService struct {
	store Store
}

func NewUserService(s Store) *UserService {
	return &UserService{store: s}
}

func (s *UserService) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users/register", s.handleUserRegister).Methods("POST")
	r.HandleFunc("/users/login", s.handleUserLogin).Methods("POST")
}

func (s *UserService) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var payload *User
	err = json.Unmarshal(body, &payload)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest,"Invalid request payload", nil)
		return
	}

	if err := validateUserPayload(payload); err != nil {
		WriteJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	hashedPassword, err := HashPassword(payload.Password)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError,  "Error creating user", nil)
		return
	}
	payload.Password = hashedPassword

	u, err := s.store.CreateUser(payload)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "Error creating user", nil)
		return
	}

	token, err := createAndSetAuthCookie(u.ID, w)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "Error creating user", nil)
		return
	}

	WriteJSON(w, http.StatusCreated,"Successful", token)
}

func (s *UserService) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var payload *User
	err = json.Unmarshal(body, &payload)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest,  "Invalid request payload", nil)
		return
	}

	if err := validateLoginUserPayload(payload); err != nil {
		WriteJSON(w, http.StatusBadRequest,  err.Error(), nil)
		
		return
	}

	user, err := s.store.GetUserByEmail(payload.Email)
	if err != nil {
		if err.Error() == "user not found" {
			WriteJSON(w, http.StatusUnauthorized,  "User not found", nil)
			
		} else {
			WriteJSON(w, http.StatusInternalServerError,"Error retrieving user", nil)
		}
		return
	}

	if !CheckPasswordHash(payload.Password, user.Password) {
		WriteJSON(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	token, err := createAndSetAuthCookie(user.ID, w)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "Error logging in", nil)
		return
	}

	WriteJSON(w, http.StatusOK,"Successful", token)
}

func validateUserPayload(user *User) error {
	if user.Email == "" {
		return errEmailRequired
	}

	if user.FirstName == "" {
		return errFirstNameRequired
	}

	if user.LastName == "" {
		return errLastNameRequired
	}

	if user.Password == "" {
		return errPasswordRequired
	}

	return nil
}

func validateLoginUserPayload(user *User) error {
	if user.Email == "" {
		return errEmailRequired
	}

	if user.Password == "" {
		return errPasswordRequired
	}

	return nil
}

func createAndSetAuthCookie(userID int64, w http.ResponseWriter) (string, error) {
	secret := []byte(Envs.JWTSecret)
	token, err := CreateJWT(secret, userID)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "Authorization",
		Value: token,
	})

	return token, nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
