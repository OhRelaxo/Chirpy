package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OhRelaxo/Chirpy/internal/auth"
	"github.com/OhRelaxo/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	log.Println("creating Chirp")
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	if err := jsonDecoder(r, &params, w); err != nil {
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("failed fetching token at <handlerPostChirps>: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "Couldn't find JWT", w)
		return
	}
	authUserId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("failed validating at <handlerPostChirps>: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "Couldn't validate JWT", w)
		return
	}

	if err := validChirp(params.Body, w); err != nil {
		return
	}

	body := filterChirps(params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   body,
		UserID: authUserId,
	})
	if err != nil {
		log.Printf("error in <handlerPostChirps> at db.CreateChirp: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
	jsonResp(http.StatusCreated, w, Chirp{
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
		log.Println("log in <validChirp>: Body too long")
		jsonErrorResp(400, "Body too long", w)
		return errors.New("")
	}
	if len(body) == 0 {
		log.Println("log in <validChirp>: Body too short or wrong parameter war used")
		jsonErrorResp(400, "Body too short or wrong parameter war used", w)
		return errors.New("")
	}
	return nil
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting Chirps")
	defer r.Body.Close()

	chirps := make([]Chirp, 0)
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("error in <handlerGetChirps> at db.GetChirps: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
	for _, dbChirp := range dbChirps {
		c := Chirp{Id: dbChirp.ID, CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.UpdatedAt, Body: dbChirp.Body, UserId: dbChirp.UserID}
		chirps = append(chirps, c)
	}
	jsonResp(http.StatusOK, w, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting a Chirp")
	defer r.Body.Close()
	pathValue := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(pathValue)
	if err != nil {
		log.Printf("error in <handlerGetChirp> at uuid.Parse: %v", err)
		jsonErrorResp(http.StatusNotFound, "please use a valid uuid", w)
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("error in <handlerGetChirp> at db.GetChirpByID: %v", err)
		jsonErrorResp(http.StatusNotFound, "please use a valid chirp id", w)
		return
	}

	chirp := Chirp{
		Id:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserId:    dbChirp.UserID,
	}

	jsonResp(http.StatusOK, w, chirp)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	log.Println("deleting chirp")

	pathValue := r.PathValue("chirpID")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error in <handlerDeleteChirp> at auth.GetBearerToken: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "unable to fetch Bearer Token", w)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("error in <handlerDeleteChirp> at auth.ValidateJWT: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "failed to validated JWT", w)
		return
	}

	chirpId, err := uuid.Parse(pathValue)
	if err != nil {
		log.Printf("error in <handlerDeleteChirp> at uuid.Parse: %v", err)
		jsonErrorResp(http.StatusBadRequest, "unable to parse uuid", w)
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpId)
	if err != nil {
		log.Printf("error in <handlerDeleteChirp> at db.GetChirpByID: %v", err)
		jsonErrorResp(http.StatusNotFound, "chirp with the given id was not found", w)
		return
	}

	if dbChirp.UserID != userId {
		log.Println("error in <handlerDeleteChirp> user does not own chirp")
		jsonErrorResp(http.StatusForbidden, "user does not own chirp", w)
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), dbChirp.ID)
	if err != nil {
		log.Printf("error in <handlerDeleteChirp> at db.DeleteChirpByID: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "failed to delete chirp in database", w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
