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
	"strconv"
	"strings"
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
		http.Error(w, "invalid access token", http.StatusInternalServerError)
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
		ID:        strconv.Itoa(cat.ID),
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

type SearchReqBody struct {
	id         string
	limit      string
	offset     string
	race       string
	sex        string
	hasMatched string
	ageInMonth string
	owned      string
	search     string
}

func (s SearchReqBody) ID() *string {
	if s.id == "" {
		return nil
	}
	return &s.id
}

func (s SearchReqBody) Limit() *int {
	if s.limit == "" {
		return nil
	}

	l, err := strconv.Atoi(s.limit)
	if err != nil {
		return nil
	}

	return &l
}

func (s SearchReqBody) Offset() *int {
	if s.offset == "" {
		return nil
	}

	o, err := strconv.Atoi(s.offset)
	if err != nil {
		return nil
	}

	return &o
}

func (s SearchReqBody) Race() *string {
	if s.race == "" {
		return nil
	}
	if !slices.Contains(races, s.race) {
		return nil
	}
	return &s.race
}

func (s SearchReqBody) Sex() *string {
	if s.race == "" {
		return nil
	}
	if s.race != "male" && s.race != "female" {
		return nil
	}
	return &s.race
}

func (s SearchReqBody) HasMatched() *bool {
	if s.hasMatched == "" {
		return nil
	}
	if s.hasMatched == "true" {
		h := true
		return &h
	}
	if s.hasMatched == "false" {
		h := false
		return &h
	}
	return nil
}

func (s SearchReqBody) AgeInMonthGreaterThan() *int {
	if s.ageInMonth == "" {
		return nil
	}
	if strings.HasPrefix(s.ageInMonth, ">") {
		a, err := strconv.Atoi(strings.TrimPrefix(s.ageInMonth, ">"))
		if err != nil {
			return nil
		}
		return &a
	}
	return nil
}

func (s SearchReqBody) AgeInMonthLessThan() *int {
	if s.ageInMonth == "" {
		return nil
	}
	if strings.HasPrefix(s.ageInMonth, "<") {
		a, err := strconv.Atoi(strings.TrimPrefix(s.ageInMonth, "<"))
		if err != nil {
			return nil
		}
		return &a
	}
	return nil
}

func (s SearchReqBody) AgeInMonth() *int {
	if s.ageInMonth == "" {
		return nil
	}
	a, err := strconv.Atoi(s.ageInMonth)
	if err != nil {
		return nil
	}
	return &a
}

func (s SearchReqBody) UserID(userID string) *string {
	if s.owned == "" {
		return nil
	}
	if s.owned == "true" {
		return &userID
	}
	return nil
}

func (s SearchReqBody) ExcludeUserID(userID string) *string {
	if s.owned == "" {
		return nil
	}
	if s.owned == "false" {
		return &userID
	}
	return nil
}

func (s SearchReqBody) NameQuery() *string {
	if s.search == "" {
		return nil
	}
	return &s.search

}
