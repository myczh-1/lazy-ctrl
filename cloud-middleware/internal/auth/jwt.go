package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/config"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"` // admin, user
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secretKey:            []byte(cfg.SecretKey),
		accessTokenDuration:  time.Duration(cfg.AccessTokenDuration) * time.Minute,
		refreshTokenDuration: time.Duration(cfg.RefreshTokenDuration) * 24 * time.Hour,
	}
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(userID, username, email, role string) (*TokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(j.accessTokenDuration)
	refreshExpiresAt := now.Add(j.refreshTokenDuration)

	// Generate access token
	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "lazy-ctrl-cloud",
			Subject:   userID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "lazy-ctrl-cloud",
			Subject:   userID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshToken generates a new access token from a valid refresh token
func (j *JWTService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// ExtractUserID extracts user ID from token string
func (j *JWTService) ExtractUserID(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// IsAdmin checks if the token belongs to an admin user
func (j *JWTService) IsAdmin(tokenString string) (bool, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return false, err
	}
	return claims.Role == "admin", nil
}