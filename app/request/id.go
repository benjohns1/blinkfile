package request

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type requestIDKey struct{}

const requestIDLength = 32

func NewID() string {
	b := make([]byte, requestIDLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func CtxWithNewID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestIDKey{}, NewID())
}

func GetID(ctx context.Context) string {
	raw := ctx.Value(requestIDKey{})
	requestID, ok := raw.(string)
	if !ok {
		return ""
	}
	return requestID
}
