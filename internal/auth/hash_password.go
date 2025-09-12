package auth

import (
	"bytes"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	sBuff := bytes.NewBufferString(password)
	crPass, err := bcrypt.GenerateFromPassword(sBuff.Bytes(), 8)
	if err != nil {
		log.Printf("error in <HashPassword>: %v", err)
		return "", err
	}
	return string(crPass), nil
}

func CheckPasswordHash(password, hash string) error {
	pBuff := bytes.NewBufferString(password)
	hBuff := bytes.NewBufferString(hash)
	err := bcrypt.CompareHashAndPassword(hBuff.Bytes(), pBuff.Bytes())
	if err != nil {
		log.Printf("error in <CheckPasswordHash>: %v", err)
		return err
	}
	return nil
}
