package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (c *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	var params parameters
	err := jsonUnmarshal(r, &params)
	if err != nil {
		log.Printf("error in <handlerCreateUser>: %v", err)
		jsonErrorResp(500, "internal server error", w)
		return
	}
	dbUser, err := c.db.CreateUser(context.Background(), params.Email)
	if err != nil {
		log.Printf("error in <handlerCreateUser>: %v", err)
		jsonErrorResp(500, "internal server error", w)
		return
	}
	resUser := User{ID: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email}

	jsonResp(201, w, resUser)
}
