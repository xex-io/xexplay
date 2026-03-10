package jwt

import (
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-key-for-unit-tests"

func makeToken(t *testing.T, claims Claims, secret string) string {
	t.Helper()
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestParse_ValidToken(t *testing.T) {
	userID := uuid.New()

	tokenStr := makeToken(t, Claims{
		UserID:    userID,
		Email:     "test@example.com",
		Role:      "user",
		TokenType: "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    "nyyu",
			Subject:   userID.String(),
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}, testSecret)

	claims, err := Parse(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", claims.Email)
	}
	if claims.Role != "user" {
		t.Errorf("Role = %v, want user", claims.Role)
	}
	if claims.TokenType != "access" {
		t.Errorf("TokenType = %v, want access", claims.TokenType)
	}
}

func TestParse_UserIDFromSubClaim(t *testing.T) {
	userID := uuid.New()

	// UserID is zero, but Subject contains valid UUID
	tokenStr := makeToken(t, Claims{
		TokenType: "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}, testSecret)

	claims, err := Parse(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v (extracted from sub)", claims.UserID, userID)
	}
}

func TestParse_ExpiredToken(t *testing.T) {
	tokenStr := makeToken(t, Claims{
		UserID:    uuid.New(),
		TokenType: "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}, testSecret)

	_, err := Parse(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestParse_InvalidSignature(t *testing.T) {
	tokenStr := makeToken(t, Claims{
		UserID:    uuid.New(),
		TokenType: "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}, testSecret)

	_, err := Parse(tokenStr, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
}

func TestParse_MissingUserID(t *testing.T) {
	// No UserID and no Subject → should fail
	tokenStr := makeToken(t, Claims{
		TokenType: "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}, testSecret)

	_, err := Parse(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for missing user_id, got nil")
	}
}

func TestParse_InvalidTokenType(t *testing.T) {
	tokenStr := makeToken(t, Claims{
		UserID:    uuid.New(),
		TokenType: "refresh",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}, testSecret)

	_, err := Parse(tokenStr, testSecret)
	if err == nil {
		t.Fatal("expected error for refresh token type, got nil")
	}
}
