package util

import (
	"errors"
	"math/rand"
	"strings"
)

func MakeRandomStr(digit uint32) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("unexpected error...")
	}

	// letters からランダムに取り出して文字列を生成
	var result strings.Builder
	for _, v := range b {
		// index が letters の長さに収まるように調整
		result.WriteString(string(letters[int(v)%len(letters)]))
	}
	return result.String(), nil
}
