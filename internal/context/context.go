package context

import (
	"context"
)

type contextKey string

const (
	ContextUserKey = contextKey("userID")
	ContextIPKey   = contextKey("ip")
)

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextUserKey).(string)
	return userID, ok
}

func GetIP(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(ContextIPKey).(string)
	return ip, ok
}
