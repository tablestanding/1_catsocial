package user

import (
	"catsocial/pkg/web"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"slices"
	"strconv"
	"strings"
)

type (
	contextKey string

	svc interface {
		Register(ctx context.Context, args RegisterArgs) (string, error)
		Login(ctx context.Context, args LoginArgs) (User, error)
		GetAccessToken(userID string) (string, error)
		IsAccessTokenValid(token string) (map[string]any, bool)
	}

	Controller struct {
		s svc
	}
)

const (
	userIDContextKey contextKey = "//user-id"
)

func NewController(s svc) Controller {
	return Controller{s}
}

type RegisterReqBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (r RegisterReqBody) Validate() bool {
	// email is not null
	if r.Email == "" {
		return false
	}

	// email should be in valid email format
	_, parseEmailErr := mail.ParseAddress(r.Email)
	if parseEmailErr != nil {
		return false
	}

	// name min length 5 and max length 50
	if len(r.Name) < 5 || len(r.Name) > 50 {
		return false
	}

	// password min length 5 and max length 15
	if len(r.Password) < 5 || len(r.Password) > 15 {
		return false
	}

	return true
}

type LoginResp struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccessToken string `json:"accessToken"`
}

func (c Controller) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := web.DecodeReqBody[RegisterReqBody](r.Body)
	if errors.Is(err, web.ErrInvalidReqBody) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := c.s.Register(r.Context(), RegisterArgs{
		Email:    reqBody.Email,
		Name:     reqBody.Name,
		Password: reqBody.Password,
	})
	if errors.Is(err, ErrUniqueEmailViolation) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := c.s.GetAccessToken(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := LoginResp{
		Email:       reqBody.Email,
		Name:        reqBody.Name,
		AccessToken: token,
	}
	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(web.NewResTemplate("User registered successfully", resp))
	if err != nil {
		http.Error(w, fmt.Sprintf("decoding user into json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}

type LoginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l LoginReqBody) Validate() bool {
	// email and name and password is not null
	if slices.Contains([]string{l.Email, l.Password}, "") {
		return false
	}

	// email should be in valid format
	_, parseEmailErr := mail.ParseAddress(l.Email)
	if parseEmailErr != nil {
		return false
	}

	// password min length 5
	if len(l.Password) < 5 {
		return false
	}

	// password max length 15
	if len(l.Password) > 15 {
		return false
	}

	return true
}

func (c Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := web.DecodeReqBody[LoginReqBody](r.Body)
	if errors.Is(err, web.ErrInvalidReqBody) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u, err := c.s.Login(r.Context(), LoginArgs{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if errors.Is(err, ErrInvalidPassword) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrUserNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := c.s.GetAccessToken(strconv.Itoa(u.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := LoginResp{
		Email:       reqBody.Email,
		Name:        u.Name,
		AccessToken: token,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(web.NewResTemplate("User logged successfully", resp))
	if err != nil {
		http.Error(w, fmt.Sprintf("decoding user into json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

func (c Controller) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")

		payload, isValid := c.s.IsAccessTokenValid(strings.TrimPrefix(a, "Bearer "))
		if !isValid {
			http.Error(w, "missing or expired access token", http.StatusUnauthorized)
			return
		}

		userID, ok := payload["userId"].(string)
		if !ok {
			http.Error(w, "missing or expired access token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}
