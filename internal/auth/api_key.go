package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	after, found := strings.CutPrefix(authHeader, "ApiKey ")
	if !found {
		return "", errors.New("no prefix \"ApiKey \" found")
	}
	return after, nil
}
