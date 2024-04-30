package web

import (
	"encoding/json"
	"fmt"
	"io"
)

type (
	Validator interface {
		Validate() bool
	}

	ResTemplate struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
)

func DecodeReqBody[Body Validator](r io.Reader) (Body, error) {
	var b Body
	err := json.NewDecoder(r).Decode(&b)
	if err != nil {
		return b, fmt.Errorf("decoding req body: %w", ErrInvalidReqBody)
	}

	valid := b.Validate()
	if !valid {
		return b, fmt.Errorf("validating req body: %w", ErrInvalidReqBody)
	}

	return b, nil
}

func NewResTemplate(msg string, data any) ResTemplate {
	return ResTemplate{Message: msg, Data: data}
}
