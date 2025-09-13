package main

import (
	"log"
	"net/http"
	"time"

	"github.com/OhRelaxo/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	log.Println("refreshing token")

	type respone struct {
		Token string `json:"token"`
	}
	refToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error in <handlerRefresh> at auth.GetBearerToken: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	dbrefToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), refToken)
	if err != nil {
		log.Printf("error in <handlerRefresh> at db.GetUserFromRefreshToken: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
	if dbrefToken.RevokedAt.Valid {
		jsonErrorResp(http.StatusUnauthorized, "you are not authorized", w)
		return
	}

	expIn, err := time.ParseDuration("1h")
	if err != nil {
		log.Printf("error in <handlerRefresh> at time.ParseDuration: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	token, err := auth.MakeJWT(dbrefToken.UserID, cfg.secret, expIn)
	if err != nil {
		log.Printf("error in <hanlderRefresh> at auth.MakeJWT: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	jsonResp(http.StatusOK, w, respone{Token: token})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error in <handlerRevoke> at auth.GetBearerToken: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refToken)
	if err != nil {
		log.Printf("error in <handlerRevoke> at db.RevokeRefreshToken: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}

	emptyResp := make([]byte, 0)
	w.WriteHeader(http.StatusNoContent)
	_, err = w.Write(emptyResp)
	if err != nil {
		log.Printf("error in <handlerRevoke> at w.Write: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "internal server error", w)
		return
	}
}
