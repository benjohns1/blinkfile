package request

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
)

type requestIDKey struct{}

const requestIDLength = 32

var ReadRandomBytes = rand.Read

var Printf = log.Printf

func NewID() string {
	b := make([]byte, requestIDLength)
	_, err := ReadRandomBytes(b)
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
		requestID = NewID()
		Printf("no request ID was attached to context, generated a new one %s", requestID)
	}
	return requestID
}
