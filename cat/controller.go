package cat

import (
	"catsocial/pkg/web"
	"catsocial/user"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"
)

type (
	svc interface {
		Create(ctx context.Context, args CreateCatArgs) (Cat, error)
	}

	Controller struct {
		s svc
	}
)

func NewController(s svc) Controller {
	return Controller{s}
}

type CreateReqBody struct {
	Race        string   `json:"race"`
	Sex         string   `json:"sex"`
	Name        string   `json:"name"`
	AgeInMonth  int      `json:"ageInMonth"`
	Description string   `json:"description"`
	ImageURLs   []string `json:"imageUrls"`
}

func (c CreateReqBody) Validate() bool {
	// name min length 1 and max length 30
	if len(c.Name) < 1 || len(c.Name) > 30 {
		return false
	}

	// must be valid race
	if !slices.Contains(races, c.Race) {
		return false
	}

	// sex is either male or female
	if c.Sex != "male" && c.Sex != "female" {
		return false
	}

	// age in month min 1 and max 120082
	if c.AgeInMonth < 1 || c.AgeInMonth > 120082 {
		return false
	}

	// description min length 1 and max length 200
	if len(c.Description) < 1 || len(c.Description) > 200 {
		return false
	}

	// imageUrls min item is 1
	if len(c.ImageURLs) < 1 {
		return false
	}

	// imageUrls should contain only valid url
	isUrl := func(str string) bool {
		u, err := url.Parse(str)
		return err == nil && u.Scheme != "" && u.Host != ""
	}
	for _, url := range c.ImageURLs {
		if !isUrl(url) {
			return false
		}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cat, err := c.s.Create(r.Context(), CreateCatArgs{
		Race:        reqBody.Race,
		Sex:         reqBody.Sex,
		Name:        reqBody.Name,
		AgeInMonth:  reqBody.AgeInMonth,
		Description: reqBody.Description,
		ImageURLs:   reqBody.ImageURLs,
		UserID:      userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateResp{
		ID:        cat.ID,
		CreatedAt: cat.CreatedAt.Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(web.NewResTemplate("success", resp))
	if err != nil {
		http.Error(w, fmt.Sprintf("decoding create cat resp into json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}
