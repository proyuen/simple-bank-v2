package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

// Maker 是管理 Token 的接口
type Maker interface {
	// CreateToken 为指定用户名创建一个新的 Token
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken 检查 Token 是否有效
	VerifyToken(token string) (*Payload, error)
}

// JWTMaker 是 JWT 的 Maker 实现
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker 创建一个新的 JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey: secretKey}, nil
}

// CreateToken 为指定用户名创建一个新的 JWT Token
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", nil, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{payload})
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", nil, err
	}

	return token, payload, nil
}

// VerifyToken 检查 Token 是否有效
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &jwtClaims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := jwtToken.Claims.(*jwtClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims.Payload, nil
}

// jwtClaims 包装 Payload 以实现 jwt.Claims 接口
type jwtClaims struct {
	*Payload
}

// GetExpirationTime 实现 jwt.Claims 接口
func (c jwtClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.ExpiredAt), nil
}

// GetIssuedAt 实现 jwt.Claims 接口
func (c jwtClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.IssuedAt), nil
}

// GetNotBefore 实现 jwt.Claims 接口
func (c jwtClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetIssuer 实现 jwt.Claims 接口
func (c jwtClaims) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject 实现 jwt.Claims 接口
func (c jwtClaims) GetSubject() (string, error) {
	return c.Username, nil
}

// GetAudience 实现 jwt.Claims 接口
func (c jwtClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
