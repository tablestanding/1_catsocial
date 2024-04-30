package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type (
	Header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}

	Payload struct {
		Exp int `json:"exp"`
	}
)

func GenerateToken(duration time.Duration, secret string) (string, error) {
	header := []byte(`{"alg":"HS256","typ":"JWT"}`)
	payload, err := json.Marshal(Payload{Exp: int(time.Now().Add(duration).Unix())})
	if err != nil {
		return "", err
	}
	headerAndPayload := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(headerAndPayload))
	signature := h.Sum(nil)

	return headerAndPayload + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func IsTokenValid(token string, secret string) bool {
	// jwt token must contain 3 elements separated by dot (.) header.payload.signature
	elems := strings.Split(token, ".")
	if len(elems) != 3 {
		return false
	}

	headerAndPayload := elems[0] + "." + elems[1]

	// hmac(header.payload) must equal to signature
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(headerAndPayload))
	signature := hash.Sum(nil)
	providedSignature, err := base64.RawURLEncoding.DecodeString(elems[2])
	if err != nil {
		return false
	}
	if !hmac.Equal(signature, providedSignature) {
		return false
	}

	// parse header
	headerStr, err := base64.RawURLEncoding.DecodeString(elems[0])
	if err != nil {
		return false
	}
	var h Header
	err = json.Unmarshal(headerStr, &h)
	if err != nil {
		return false
	}

	// currently we only accept jwt token with alg HMAC SHA256 and type JWT
	if h.Alg != "HS256" || h.Typ != "JWT" {
		return false
	}

	// parse payload
	payloadStr, err := base64.RawURLEncoding.DecodeString(elems[1])
	if err != nil {
		return false
	}
	var p Payload
	err = json.Unmarshal(payloadStr, &p)
	if err != nil {
		return false
	}

	// check if payload exp already expires
	exp := time.Unix(int64(p.Exp), 0)
	if time.Now().After(exp) {
		return false
	}

	return true
}
