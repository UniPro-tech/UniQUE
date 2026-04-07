package util

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword bcryptでハッシュ化して返す
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword bcryptのハッシュと照合する
func VerifyPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	// bcrypt malformed hash: try legacy SHA256 hex(password)
	if _, ok := err.(bcrypt.InvalidHashPrefixError); ok {
		sum := sha256.Sum256([]byte(password))
		if hex.EncodeToString(sum[:]) == hash {
			return true, nil
		}
		return false, nil
	}
	// other bcrypt error
	return false, err
}
