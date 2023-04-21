package grpcserver

import (
	"context"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryUserIDInterceptor fetch hashed user id from request context and replaces it with decrypted user id
func UnaryUserIDInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var token string
	userID := uuid.Nil
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(middlewares.CookieName)
		if len(values) > 0 {
			token = values[0]
			userID, _ = middlewares.DecodeUserIDFromHashedString(token)
		}
	}
	if userID == uuid.Nil {
		userID = uuid.New()
	}
	newCtx := context.WithValue(ctx, middlewares.UserIDKey, userID)
	return handler(newCtx, req)
}
