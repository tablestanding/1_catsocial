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
)

type (
	svc interface {
		Register(ctx context.Context, args RegisterArgs) error
		Login(ctx context.Context, args LoginArgs) (User, error)
		GetAccessToken(ctx context.Context) (string, error)
	}

	Controller struct {
		s svc
	}
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
	// email and name and password is not null
	if slices.Contains([]string{r.Email, r.Name, r.Password}, "") {
		return false
	}

	// email should be in valid format
	_, parseEmailErr := mail.ParseAddress(r.Email)
	if parseEmailErr != nil {
		return false
	}

	// name min length 5
	if len(r.Name) < 5 {
		return false
	}

	// name max length 50
	if len(r.Name) > 50 {
		return false
	}

	// password min length 5
	if len(r.Password) < 5 {
		return false
	}

	// password max length 15
	if len(r.Password) > 15 {
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

	err = c.s.Register(r.Context(), RegisterArgs{
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

	token, err := c.s.GetAccessToken(r.Context())
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
	err = json.NewEncoder(w).Encode(web.NewResTemplate("User registered successfully", resp))
	if err != nil {
		http.Error(w, fmt.Sprintf("decoding user into json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
