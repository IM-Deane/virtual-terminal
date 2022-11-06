package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// writeJSON creates a JSON response from the supplied parameters
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		// set headers
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	// send back JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil
}

// readJSON helper that reads JSON from request body
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct {}{})
	if err != io.EOF {
		return errors.New("body must only have single JSON value")
	}

	return nil
}

// badRequest returns a formatted error response as JSON
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct{
		Error bool `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(out)
	return nil
}

// invalidCredentials sends back an error message for bad login credentials
func (app *application) invalidCredentials(w http.ResponseWriter) error {
	var payload struct {
		Error bool `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = "Invalid authentication credentials"

	err := app.writeJSON(w, http.StatusUnauthorized, payload)
	if err != nil {
		return err
	}

	return nil
}


// passwordMatches validates the provided password
func (app *application) passwordMatches(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}