package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func jsonErrorResp(code int, msg string, w http.ResponseWriter) {
	type jsonError struct {
		Error string `json:"error"`
	}
	log.Printf("Responding with %v error: %v", code, msg)
	jsonResp(code, w, jsonError{Error: msg})
}

func jsonResp(code int, w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	rawJson, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error in <jsonResp> at json.Marshal: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(rawJson)

}
