package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

const invalidTokenPrefix = "invalid_token:"

// invalidateToken records the given JWT in Redis so it is rejected until it naturally expires.
// It is a no-op when no cache is configured or the token is already expired.
func (s *Service) invalidateToken(tokenString string) error {
	if s.cache == nil || tokenString == "" {
		return nil
	}

	token, err := web.ParseJWT(s.jwtKey, tokenString,
		jwt.WithIssuer(s.jwtIssuer),
		jwt.WithAudience(s.jwtAudience),
	)
	if err != nil || !token.Valid {
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return nil
	}

	ttl := time.Until(exp.Time)
	if ttl <= 0 {
		return nil // already expired — nothing to invalidate
	}

	key := fmt.Sprintf("%s%s", invalidTokenPrefix, tokenString)
	if s.cache != nil {
		return s.cache.SetEX(context.Background(), key, "1", ttl).Err()
	}

	return fmt.Errorf("no cache to invalidate token")
}

// isTokenInvalidated returns true if the token was explicitly invalidated (e.g. on logout).
func (s *Service) isTokenInvalidated(tokenString string) bool {
	if s.cache == nil || tokenString == "" {
		return false
	}

	key := fmt.Sprintf("%s%s", invalidTokenPrefix, tokenString)
	exists, err := s.cache.Exists(context.Background(), key).Result()
	if err != nil {
		return false
	}

	return exists > 0
}
