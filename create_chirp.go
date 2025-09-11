package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OhRelaxo/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	log.Println("creating Chirp")
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	params := parameters{}
	if err := jsonDecoder(r, &params, w); err != nil {
		return
	}

	if err := validChirp(params.Body, w); err != nil {
		return
	}

	body := filterChirps(params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   body,
		UserID: params.UserId,
	})
	if err != nil {
		log.Printf("error in <handlerChirps> at db.CreateChirp: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
	jsonResp(http.StatusCreated, w, returnVals{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func filterChirps(body string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	reqString := strings.Fields(body)
	for i, s := range reqString {
		if _, ok := badWords[strings.ToLower(s)]; ok {
			reqString[i] = "****"
		}
	}
	return strings.Join(reqString, " ")
}

func validChirp(body string, w http.ResponseWriter) error {
	if len(body) > 140 {
		log.Println("log in <handlerChirps>: Body too long")
		jsonErrorResp(400, "Body too long", w)
		return errors.New("")
	}
	if len(body) == 0 {
		log.Println("log in <handlerChirps>: Body too short or wrong parameter war used")
		jsonErrorResp(400, "Body too short or wrong parameter war used", w)
		return errors.New("")
	}
	return nil
}
