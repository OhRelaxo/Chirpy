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

func jsonResp(code int, w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	rawJson, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error in <jsonResp> at json.Marshal: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	_, err = w.Write(rawJson)
	if err != nil {
		return
	}
}

func jsonDecoder(r *http.Request, params any, w http.ResponseWriter) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("error in <jsonDecoder>: at decoder.Decode: %v", err)
		jsonErrorResp(500, "internal server error", w)
		return err
	}
	return nil
}
