package tokens

import (
	"errors"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"tarkib.uz/config"
)

// JWTHandler ...
type JWTHandler struct {
	Sub       string
	Iss       string
	Exp       string
	Iat       string
	Aud       []string
	Role      string
	SigninKey string
	Log       *log.Logger
	Token     string
	Timeout   int
}

type CustomClaims struct {
	*jwt.Token
	Sub  string   `json:"sub"`
	Exp  float64  `json:"exp"`
	Iat  float64  `json:"iat"`
	Aud  []string `json:"aud"`
	Role string   `json:"role"`
}

// GenerateAuthJWT ...
func (jwtHandler *JWTHandler) GenerateAuthJWT() (access, refresh string, err error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return "", "", err
	}

	var (
		accessToken  *jwt.Token
		refreshToken *jwt.Token
		claims       jwt.MapClaims
		rtClaims     jwt.MapClaims
	)

	accessToken = jwt.New(jwt.SigningMethodHS256)
	refreshToken = jwt.New(jwt.SigningMethodHS256)
	claims = accessToken.Claims.(jwt.MapClaims)
	claims["sub"] = jwtHandler.Sub
	claims["exp"] = jwtHandler.Exp
	claims["iat"] = time.Now().Unix()
	claims["role"] = jwtHandler.Role
	claims["aud"] = jwtHandler.Aud
	access, err = accessToken.SignedString([]byte(cfg.Casbin.SigningKey))
	if err != nil {
		log.Println("error generating access token", err)
		return
	}

	rtClaims = refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = jwtHandler.Sub
	refresh, err = refreshToken.SignedString([]byte(jwtHandler.SigninKey))
	if err != nil {
		log.Println("error generating refresh token", err)
		return
	}
	return
}

// ExtractClaims ...
func (jwtHandler *JWTHandler) ExtractClaims() (jwt.MapClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	// Parse the token
	token, err = jwt.Parse(jwtHandler.Token, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtHandler.SigninKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Verify token validity and extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		log.Println("invalid jwt token")
		return nil, err
	}

	// Check token expiration
	expStr, ok := claims["exp"].(string) // The exp claim is in ISO 8601 format as a string
	if !ok {
		log.Println("expiration claim is missing or invalid")
		return nil, errors.New("expiration claim is missing or invalid")
	}

	expTime, err := time.Parse(time.RFC3339, expStr)
	if err != nil {
		log.Println("error parsing expiration time:", err)
		return nil, err
	}

	if time.Now().After(expTime) {
		log.Println("token is expired")
		return nil, errors.New("token is expired huh")
	}

	return claims, nil
}
