package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the XEX Exchange JWT claims structure.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

// Sign creates a new JWT token with the given claims, signed with the shared secret.
func Sign(claims *Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Parse validates and parses an Exchange-issued JWT using the shared secret.
func Parse(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.TokenType != "" && claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: %s", claims.TokenType)
	}

	// Extract user_id from sub claim if UserID is zero
	if claims.UserID == uuid.Nil && claims.Subject != "" {
		parsed, err := uuid.Parse(claims.Subject)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id in sub claim: %w", err)
		}
		claims.UserID = parsed
	}

	if claims.UserID == uuid.Nil {
		return nil, fmt.Errorf("missing user_id in token")
	}

	return claims, nil
}
