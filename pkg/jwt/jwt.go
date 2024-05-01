package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type (
	Header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
)

func GenerateToken(duration time.Duration, secret string, p map[string]any) (string, error) {
	header := []byte(`{"alg":"HS256","typ":"JWT"}`)
	p["exp"] = int(time.Now().Add(duration).Unix())
	payload, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("jwt generate token: %w", err)
	}
	headerAndPayload := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(headerAndPayload))
	signature := h.Sum(nil)

	return headerAndPayload + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

// IsTokenValid returns the token payload and boolean that will be true if the token is valid
func IsTokenValid(token string, secret string) (map[string]any, bool) {
	// jwt token must contain 3 elements separated by dot (.) header.payload.signature
	elems := strings.Split(token, ".")
	if len(elems) != 3 {
		return nil, false
	}

	headerAndPayload := elems[0] + "." + elems[1]

	// hmac(header.payload) must equal to signature
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(headerAndPayload))
	signature := hash.Sum(nil)
	providedSignature, err := base64.RawURLEncoding.DecodeString(elems[2])
	if err != nil {
		return nil, false
	}
	if !hmac.Equal(signature, providedSignature) {
		return nil, false
	}

	// parse header
	headerStr, err := base64.RawURLEncoding.DecodeString(elems[0])
	if err != nil {
		return nil, false
	}
	var h Header
	err = json.Unmarshal(headerStr, &h)
	if err != nil {
		return nil, false
	}

	// currently we only accept jwt token with alg HMAC SHA256 and type JWT
	if h.Alg != "HS256" || h.Typ != "JWT" {
		return nil, false
	}

	// parse payload
	payloadStr, err := base64.RawURLEncoding.DecodeString(elems[1])
	if err != nil {
		return nil, false
	}
	var p map[string]any
	err = json.Unmarshal(payloadStr, &p)
	if err != nil {
		return nil, false
	}

	// check if payload exp already expires
	exp, ok := p["exp"].(float64)
	if !ok {
		return nil, false
	}
	expTime := time.Unix(int64(exp), 0)
	if time.Now().After(expTime) {
		return nil, false
	}

	return p, true
}
