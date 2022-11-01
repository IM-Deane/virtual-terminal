package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IM-Deane/virtual-terminal/internal/cards"
	"github.com/go-chi/chi/v5"
)


type stripePayload struct {
	Currency string `json:"currency"`
	Amount string `json:"amount"`
}

type jsonResponse struct {
	OK bool `json:"ok"`
	// "omitempty": if empty we don't bother converting to json
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID int `json:"id"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// convert amount to string
	amount , err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card {
		Secret: app.config.stripe.secret,
		Key: app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true

	// connect to stripe
	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	// if there was no error with payment
	if okay {
		out, err := json.MarshalIndent(pi, "", "  ")
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		
		// send back success response
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		// something went wrong with payment
		j := jsonResponse{
			OK: false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "   ")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}
}

func (app *application) GetWidgetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	out, err := json.MarshalIndent(widget, "", "   ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// write out result
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}