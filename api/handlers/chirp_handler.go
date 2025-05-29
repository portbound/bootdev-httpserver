package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"
)

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		Valid bool   `json:"valid"`
		Body  string `json:"body"`
		Error string `json:"error,omitempty"`
	}

	req := requestBody{}
	res := responseBody{
		Valid: true,
		Body:  req.Body,
		Error: "",
	}
	decoder := json.NewDecoder(r.Body)

	code := http.StatusOK

	err := decoder.Decode(&req)
	if err != nil {
		code = http.StatusBadRequest
		res.Valid = false
		res.Error = "Something went wrong"
	}

	if utf8.RuneCountInString(req.Body) > 140 {
		code = http.StatusBadRequest
		res.Valid = false
		res.Error = "Chirp is too long"
	}

	res.Body = cleanChirp(req.Body)

	dat, err := json.Marshal(res)
	if err != nil {
		code = http.StatusBadRequest
		res.Valid = false
		res.Error = "Something went wrong"
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(dat))
}

func cleanChirp(s string) string {
	s = strings.ToLower(s)

	s = strings.ReplaceAll(s, "kerfuffle", "****")
	s = strings.ReplaceAll(s, "sharbert", "****")
	s = strings.ReplaceAll(s, "fornax", "****")

	return s
}
