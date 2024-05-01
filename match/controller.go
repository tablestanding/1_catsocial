package match

import (
	"catsocial/cat"
	"catsocial/pkg/web"
	"catsocial/user"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type (
	svc interface {
		Create(ctx context.Context, args CreateArgs) error
		Get(ctx context.Context, args GetArgs) ([]Match, error)
		Approve(ctx context.Context, args ApproveArgs) error
	}

	Controller struct {
		s svc
	}
)

func NewController(s svc) Controller {
	return Controller{s}
}

type CreateReqBody struct {
	MatchCatID string `json:"matchCatId"`
	UserCatID  string `json:"userCatId"`
	Msg        string `json:"message"`
}

func (c CreateReqBody) Validate() bool {
	// msg min length 5 and max length 120
	if len(c.Msg) < 5 || len(c.Msg) > 120 {
		return false
	}

	// match and user cat id must not be empty
	if c.MatchCatID == "" || c.UserCatID == "" {
		return false
	}

	// match cat id must be valid id
	_, err := strconv.Atoi(c.MatchCatID)
	if err != nil {
		return false
	}

	// user cat id must be valid id
	_, err = strconv.Atoi(c.UserCatID)
	if err != nil {
		return false
	}

	return true
}

type CreateResp struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

func (c Controller) CreateHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := web.DecodeReqBody[CreateReqBody](r.Body)
	if errors.Is(err, web.ErrInvalidReqBody) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, ok := user.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "invalid access token", http.StatusInternalServerError)
		return
	}

	err = c.s.Create(r.Context(), CreateArgs{
		MatchCatID: reqBody.MatchCatID,
		UserCatID:  reqBody.UserCatID,
		UserID:     userID,
		Msg:        reqBody.Msg,
	})
	if errors.Is(err, cat.ErrCatNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if errors.Is(err, ErrCatHasBeenMatched) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrCatsFromTheSameOwner) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrCatsHaveSameGender) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrUserDoesNotOwnCat) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type (
	GetRespItem struct {
		ID             string              `json:"id"`
		Msg            string              `json:"message"`
		CreatedAt      string              `json:"createdAt"`
		IssuedBy       GetRespItemIssuedBy `json:"issuedBy"`
		MatchCatDetail cat.SearchRespItem  `json:"matchCatDetail"`
		UserCatDetail  cat.SearchRespItem  `json:"userCatDetail"`
	}

	GetRespItemIssuedBy struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"createdAt"`
	}
)

func (c Controller) GetHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := user.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "invalid access token", http.StatusInternalServerError)
		return
	}

	matches, err := c.s.Get(r.Context(), GetArgs{UserID: userID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var items []GetRespItem
	for _, m := range matches {
		userCat := m.ReceiverCat
		matchCat := m.IssuerCat
		if strconv.Itoa(m.IssuerUser.ID) == userID {
			userCat = m.IssuerCat
			matchCat = m.ReceiverCat
		}
		items = append(items, GetRespItem{
			ID:        strconv.Itoa(m.ID),
			Msg:       m.Msg,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
			IssuedBy: GetRespItemIssuedBy{
				Email:     m.IssuerUser.Email,
				Name:      m.IssuerUser.Name,
				CreatedAt: m.IssuerUser.CreatedAt.Format(time.RFC3339),
			},
			UserCatDetail: cat.SearchRespItem{
				ID:          strconv.Itoa(userCat.ID),
				Name:        userCat.Name,
				Race:        userCat.Race,
				Sex:         userCat.Sex,
				AgeInMonth:  userCat.AgeInMonth,
				ImageURLs:   userCat.ImageURLs,
				Description: userCat.Description,
				HasMatched:  userCat.HasMatched,
				CreatedAt:   userCat.CreatedAt.Format(time.RFC3339),
			},
			MatchCatDetail: cat.SearchRespItem{
				ID:          strconv.Itoa(matchCat.ID),
				Name:        matchCat.Name,
				Race:        matchCat.Race,
				Sex:         matchCat.Sex,
				AgeInMonth:  matchCat.AgeInMonth,
				ImageURLs:   matchCat.ImageURLs,
				Description: matchCat.Description,
				HasMatched:  matchCat.HasMatched,
				CreatedAt:   matchCat.CreatedAt.Format(time.RFC3339),
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(web.NewResTemplate("success", items))
	if err != nil {
		http.Error(w, fmt.Sprintf("decoding matches into json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

type ApproveReqBody struct {
	MatchID string `json:"matchId"`
}

func (a ApproveReqBody) Validate() bool {
	// match id must not be empty
	if a.MatchID == "" {
		return false
	}

	// match id must be valid id
	_, err := strconv.Atoi(a.MatchID)
	if err != nil {
		return false
	}

	return true
}

func (c Controller) ApproveHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := web.DecodeReqBody[ApproveReqBody](r.Body)
	if errors.Is(err, web.ErrInvalidReqBody) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, ok := user.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "invalid access token", http.StatusInternalServerError)
		return
	}

	err = c.s.Approve(r.Context(), ApproveArgs{
		ID: reqBody.MatchID,
	})
	if errors.Is(err, ErrMatchNotValid) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrMatchNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
