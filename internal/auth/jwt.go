package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID      uuid.UUID   `json:"user_id"`
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	Roles       []string    `json:"roles"`
	Permissions []string    `json:"permissions"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey          []byte
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
	issuer             string
}

func NewJWTService(secret string, accessTokenTTL, refreshTokenTTL time.Duration, issuer string) *JWTService {
	return &JWTService{
		secretKey:       []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		issuer:          issuer,
	}
}

// GenerateAccessToken generates a new JWT access token
func (j *JWTService) GenerateAccessToken(userID uuid.UUID, username, email string, roles, permissions []string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    j.issuer,
			Subject:   userID.String(),
			Audience:  []string{"pustaka"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken generates a new JWT refresh token
func (j *JWTService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    j.issuer,
		Subject:   userID.String(),
		Audience:  []string{"pustaka-refresh"},
		ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenTTL)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != j.issuer {
		return nil, errors.New("invalid token issuer")
	}

	// Validate audience
	if !verifyAudience(claims.Audience, "pustaka") {
		return nil, errors.New("invalid token audience")
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses a refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, jwt.ErrTokenExpired
		}
		return uuid.Nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid refresh token claims")
	}

	// Validate issuer
	if claims.Issuer != j.issuer {
		return uuid.Nil, errors.New("invalid token issuer")
	}

	// Validate audience
	if !verifyAudience(claims.Audience, "pustaka-refresh") {
		return uuid.Nil, errors.New("invalid token audience")
	}

	// Extract user ID
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}

// RefreshToken generates a new access token from a refresh token
func (j *JWTService) RefreshToken(refreshTokenString string, username, email string, roles, permissions []string) (string, string, error) {
	userID, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new tokens
	newAccessToken, err := j.GenerateAccessToken(userID, username, email, roles, permissions)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := j.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must be in format 'Bearer {token}'")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetTokenExpiration returns the expiration time of a token
func (j *JWTService) GetTokenExpiration(tokenString string) (time.Time, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return time.Time{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return time.Time{}, errors.New("invalid token claims")
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("token has no expiration time")
	}

	return claims.ExpiresAt.Time, nil
}

// verifyAudience checks if the expected audience is in the list of audiences
func verifyAudience(audiences []string, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}

// GetAccessTokenTTL returns the access token TTL
func (j *JWTService) GetAccessTokenTTL() time.Duration {
	return j.accessTokenTTL
}