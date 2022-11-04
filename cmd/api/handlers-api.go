package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/IM-Deane/virtual-terminal/internal/cards"
	"github.com/IM-Deane/virtual-terminal/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v73"
)


type stripePayload struct {
	Currency string `json:"currency"`
	Amount string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email string `json:"email"`
	CardBrand string `json:"card_brand"`
	LastFour string `json:"last_four"`
	PlanID string `json:"plan"`
	ExpiryMonth int `json:"exp_month"`
	ExpiryYear int `json:"exp_year"`
	ProductID string `json:"product_id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
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

func (app *application) CreateCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var data stripePayload
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card {
		Secret: app.config.stripe.secret,
		Key: app.config.stripe.key,
		Currency: data.Currency,
	}

	okay := true
	var subscription *stripe.Subscription
	txnMsg := "Transaction Successful"

	// create customer
	stripeCustomer, msg, err := card.CreateCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.errorLog.Println(err)
		okay = false
		txnMsg = msg
	}

	if okay {
		// create new subscription for customer
		subscription, err = card.SubscribeToPlan(stripeCustomer, data.PlanID, data.Email, data.LastFour, "")
		if err != nil {
			app.errorLog.Println(err)
			okay = false
			txnMsg = "Error while subscribing to plan"
		}
		app.infoLog.Println("subscription id is", subscription.ID)
	}

	if okay {
		productID, err := strconv.Atoi(data.ProductID)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		// save customer data to DB
		customerID, err := app.SaveCustomer(data.FirstName, data.LastName, data.Email)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		// create new transaction
		amount, err := strconv.Atoi(data.Amount)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		txn := models.Transaction{
			Amount: amount,
			Currency: "cad",
			LastFour: data.LastFour,
			ExpiryMonth: data.ExpiryMonth,
			ExpiryYear: data.ExpiryYear,
			TransactionStatusID: 2, // cleared
		}
		txnID, err := app.SaveTransaction(txn)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		// create order
		order := models.Order{
			WidgetID: productID,
			TransactionID: txnID,
			CustomerID: customerID,
			StatusID: 1, // cleared
			Quantity: 1,
			Amount: amount,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err = app.SaveOrder(order)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	resp := jsonResponse{
		OK: okay,
		Message: txnMsg,
	}

	out, err := json.MarshalIndent(resp, "", "   ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// SaveCustomer handler that saves a customer and returns its id
func (app *application) SaveCustomer(firstName string, lastName string, email string) (id int, err error) {
	// business logic
	customer := models.Customer{
		FirstName: firstName,
		LastName: lastName,
		Email: email,
	}
	// hit database
	id, err = app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return id, nil
}


// SaveTransaction handler that saves a transaction and returns its id
func (app *application) SaveTransaction(txn models.Transaction) (id int, err error) {
	id, err = app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SaveOrder handler that saves a order and returns its id
func (app *application) SaveOrder(order models.Order) (id int, err error) {
	id, err = app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}