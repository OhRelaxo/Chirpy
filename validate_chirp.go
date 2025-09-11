package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("error in <handlerValidateChirp>: at decoder.Decode: %v", err)
		jsonErrorResp(500, "internal server error", w)
		return
	}

	if len(params.Body) > 140 {
		log.Println("log in <handlerValidateChrip>: Body too long")
		jsonErrorResp(400, "Body too long", w)
		return
	}
	if len(params.Body) == 0 {
		log.Println("log in <handlerValidateChrip>: Body too short")
		jsonErrorResp(400, "Body too long or wrong parameter war used", w)
		return
	}
	jsonResp(200, w, jsonValid{Valid: true})
}
