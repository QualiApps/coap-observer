package utils

import (
	"math/rand"
	"strings"
	"time"
)

var (
	chars        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	CurMessageID uint16
)

func GenToken(length int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	token := make([]rune, length)
	charsLen := len(chars)
	for i := range token {
		token[i] = chars[rand.Intn(charsLen)]
	}
	return string(token)
}

func GenMessageID() uint16 {
	if CurMessageID == 65535 {
		CurMessageID = 0
	} else {
		CurMessageID++
	}
	return CurMessageID
}

func IsEmpty(s string) bool {
	empty := false
	if len(strings.TrimSpace(s)) == 0 {
		empty = true
	}

	return empty
}
