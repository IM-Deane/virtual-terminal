package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/IM-Deane/virtual-terminal/internal/cards"
	"github.com/IM-Deane/virtual-terminal/internal/models"
	"github.com/go-chi/chi/v5"
)

// Home displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminal displays the home page
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// PaymentSucceeded displays the recipet page after purchasing a widget
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read submitted form data
	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	email := r.Form.Get("cardholder-email")
	paymentIntent := r.Form.Get("payment-intent")
	paymentMethod := r.Form.Get("payment-method")
	paymentAmount := r.Form.Get("payment-amount")
	paymentCurrency := r.Form.Get("payment-currency")

	widgetID, err := strconv.Atoi(r.Form.Get("product-id"))
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key: app.config.stripe.key,
	}

	// get payment intent
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// get payment method
	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// get card details
	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// create new customer
	customerID, err := app.SaveCustomer(firstName, lastName, email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// convert string to int
	amount, err := strconv.Atoi(paymentAmount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new transaction
	txn := models.Transaction{
		Amount: amount,
		Currency: paymentCurrency,
		LastFour: lastFour,
		ExpiryMonth: int(expiryMonth),
		ExpiryYear: int(expiryYear),
		BankReturnCode: pi.Charges.Data[0].ID,
		TransactionStatusID: 2, // transaction_status = "cleared"
	}
	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new order
	order := models.Order{
		WidgetID: widgetID,
		TransactionID: txnID,
		CustomerID: customerID,
		StatusID: 1, // statuses.cleared
		Quantity: 1, // TODO: can only buy one for nows
		Amount: amount,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}


	data := make(map[string]interface{})
	data["email"] = email
	data["first_name"] = firstName
	data["last_name"] = lastName
	data["payment_intent"] = paymentIntent
	data["payment_method"] = paymentMethod
	data["payment_amount"] = paymentAmount
	data["payment_currency"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	data["bank_return_code"] = pi.Charges.Data[0].ID

	// TODO: should write this data to session andd redirect user to new page
	// this avoids duplicate form submissions

	if err := app.renderTemplate(w, r, "succeeded", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
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

// ChargeOnce displays the page to buy a widget
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
