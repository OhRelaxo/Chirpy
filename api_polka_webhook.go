package main

import (
	"log"
	"net/http"

	"github.com/OhRelaxo/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("handling Polka Webhook")
	defer r.Body.Close()
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("error in <handlerPolkaWebhook> at auth.GetAPIKey: %v", err)
		jsonErrorResp(http.StatusUnauthorized, "no prefix \"ApiKey \" was found in the header", w)
		return
	}
	if apiKey != cfg.polkaKey {
		jsonErrorResp(http.StatusForbidden, "wrong key", w)
		return
	}

	params := parameters{}
	err = jsonDecoder(r, &params, w)
	if err != nil {
		log.Printf("error in <hanlderPolkaWebhook> at jsonDecoder: %v", err)
		jsonErrorResp(http.StatusInternalServerError, "failed to decode json", w)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userId, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		log.Printf("error in <handlerPolkaWebhook> at uuid.Parse: %v", err)
		jsonErrorResp(http.StatusNotFound, "please use a valid uuid", w)
		return
	}
	err = cfg.db.UpgradeToChirpyRed(r.Context(), userId)
	if err != nil {
		log.Printf("error in <hanlderPolkaWebhook> at db.UpgradeToChirpyRed: %v", err)
		jsonErrorResp(http.StatusNotFound, "user with given userid was not found", w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
