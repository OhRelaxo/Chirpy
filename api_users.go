package main

import (
	"log"
	"net/http"
	"time"

	"github.com/OhRelaxo/Chirpy/internal/auth"
	"github.com/OhRelaxo/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
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
		Email    string `json:"email"`
		Password string `json:"password"`
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

	expInJWT, err := time.ParseDuration("1h")
	if err != nil {
		log.Printf("error in <handlerLogin> at time.ParseDuration: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, expInJWT)
	if err != nil {
		log.Printf("error in <handlerLogin at auth.MakeJWT: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	refreshDur, err := time.ParseDuration("1440h")
	if err != nil {
		log.Printf("error in <handlerLogin> at time.ParseDuration: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
	expInRef := time.Now().Add(refreshDur)

	refTokenStr, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("error in <handlerLogin> at auth.MakeRefreshToken: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	_, err = cfg.db.CreateToken(r.Context(), database.CreateTokenParams{refTokenStr, dbUser.ID, expInRef})
	if err != nil {
		log.Printf("error in <handlerLogin> at db.CreateToken: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	jsonResp(http.StatusOK, w, User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refTokenStr,
	})
}

func (cfg *apiConfig) handlerUpdateLoginDetails(w http.ResponseWriter, r *http.Request) {
	log.Println("updating login details")
	defer r.Body.Close()
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error in <handlerUpdateLoginDetails> at auth.GetBearerToken: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "unable to fetch Bearer Token", w)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("error in <handlerUpdateLoginDetails> at auth.ValidateJWT: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "failed to validated JWT", w)
		return
	}

	params := parameters{}
	err = jsonDecoder(r, &params, w)
	if err != nil {
		log.Printf("error in <handlerUpdateLoginDetails> at jsonDecoder: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "failed to decode json", w)
		return
	}

	hashPass, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("error in <handlerUpdateLoginDetails> at auth.HashPassword: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "failed to hash password", w)
		return
	}

	dbUser, err := cfg.db.UpdateLoginDetails(r.Context(), database.UpdateLoginDetailsParams{Email: params.Email, HashedPassword: hashPass, ID: userId})
	if err != nil {
		log.Printf("error in <handlerUpdateLoginDetails> at db.UpdateLoginDetails: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "failed to updated database", w)
		return
	}

	jsonResp(http.StatusOK, w, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}
