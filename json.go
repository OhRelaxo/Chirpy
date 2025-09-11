package main

import (
	"encoding/json"
	"io"
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
	w.Write(rawJson)
}

func jsonUnmarshal(r *http.Request, parameters any) error {
	defer r.Body.Close()
	req, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error in <jsonUnmarshal> at io.ReadAll: %v", err)
		return err
	}
	json.Unmarshal(req, &parameters)
	return nil
}
