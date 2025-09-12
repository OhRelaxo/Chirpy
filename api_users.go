package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OhRelaxo/Chirpy/internal/auth"
	"github.com/OhRelaxo/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var params parameters
	if err := jsonDecoder(r, &params, w); err != nil {
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashedPassword})
	if err != nil {
		log.Printf("error in <handlerCreateUser>: %v", err)
		jsonErrorResp(500, "internal server error", w)
		return
	}
	resUser := User{ID: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email}

	jsonResp(201, w, resUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
	}
	params := parameters{}
	if err := jsonDecoder(r, &params, w); err != nil {
		return
	}
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("email not found in <handlerLogin>")
		jsonErrorResp(http.StatusUnauthorized, "Incorrect email or password", w)
		return
	}
	if err = auth.CheckPasswordHash(params.Password, dbUser.HashedPassword); err != nil {
		log.Printf("wrong password in <handlerLogin>")
		jsonErrorResp(http.StatusUnauthorized, "Incorrect email or password", w)
		return
	}

	var expIn time.Duration
	defaultDuration, err := time.ParseDuration("1h")
	if err != nil {
		fmt.Printf("error in <handlerLogin> at time.ParseDuration: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	if params.ExpiresInSeconds == nil || *params.ExpiresInSeconds > 3600 {
		expIn = defaultDuration
	} else {
		expIn, err = time.ParseDuration(fmt.Sprint(params.ExpiresInSeconds))
		if err != nil {
			fmt.Printf("error in <handlerLogin> at time.ParseDuration: %v", err)
			jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
			return
		}
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, expIn)
	if err != nil {
		log.Printf("error in <handlerLogin at auth.MakeJWT: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	jsonResp(http.StatusOK, w, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     token,
	})
}
