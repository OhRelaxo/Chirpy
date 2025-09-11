package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	reqString := strings.Fields(params.Body)
	for i, s := range reqString {
		if _, ok := badWords[strings.ToLower(s)]; ok {
			reqString[i] = "****"
		}
	}
	resString := strings.Join(reqString, " ")

	jsonResp(200, w, returnVals{CleanedBody: resString})
}
