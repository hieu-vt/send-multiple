package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type JWT struct {
	privateKey []byte
	publicKey  []byte
}

func NewJWT(privateKey []byte, publicKey []byte) JWT {
	return JWT{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (j JWT) Validate(c *gin.Context) (interface{}, error) {
	var auth string
	if len(c.GetHeader("Authorization")) > 0 {
		auth = c.GetHeader("Authorization")
	}
	if len(c.Query("token")) > 0 {
		auth = c.Query("token")
	}
	if len(c.Query("access_token")) > 0 {
		auth = c.Query("access_token")
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	key, err := jwt.ParseRSAPublicKeyFromPEM(j.publicKey)
	if err != nil {
		return "", fmt.Errorf("validate: parse key: %w", err)
	}
	tok, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		log.Printf("Validate token: %s, %v", token, err)
		return nil, fmt.Errorf("validate: %w", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("validate: invalid")
	}
	return claims, nil
}
