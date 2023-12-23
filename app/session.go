package app

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type (
	Token string

	Session struct {
		Token
		Username string
		Expires  time.Time
	}
)

func (s *Session) isValid(now func() time.Time) bool {
	return s.Expires.After(now())
}

func generateDefaultToken() (Token, error) {
	const tokenLength = 128
	v, err := randomBase64String(tokenLength)
	return Token(v), err
}

func randomBase64String(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
