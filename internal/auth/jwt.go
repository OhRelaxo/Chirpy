package auth

import (
	"bytes"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	sUserId := fmt.Sprint(userID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "Chirpy",
		IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(expiresIn)},
		Subject:   sUserId,
	})
	buffTokenSecret := bytes.NewBufferString(tokenSecret)
	signedToken, err := token.SignedString(buffTokenSecret.Bytes())
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return [16]byte{}, err
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		return [16]byte{}, err
	}
	userId, err := uuid.Parse(id)
	if err != nil {
		return [16]byte{}, err
	}
	return userId, nil
}
