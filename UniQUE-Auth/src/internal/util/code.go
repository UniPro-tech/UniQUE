package util

import (
	"encoding/base64"

	"github.com/google/uuid"
)

func GenerateAuthCode() (string, error) {
	code, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString([]byte(code.String())), nil
}
