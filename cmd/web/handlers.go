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

type TransactionData struct {
	FirstName string
	LastName string
	Email string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount int
	PaymentCurrency string
	LastFour string
	ExpiryMonth int
	ExpiryYear int
	BankReturnCode string
}

// GetTransactionData gets txn data from POST (ie. charged card) and Stripe
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// TODO: should validate data first

	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	email := r.Form.Get("cardholder-email")
	paymentIntent := r.Form.Get("payment-intent")
	paymentMethod := r.Form.Get("payment-method")
	paymentAmount := r.Form.Get("payment-amount")
	paymentCurrency := r.Form.Get("payment-currency")
	// convert string to int
	amount, err := strconv.Atoi(paymentAmount)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key: app.config.stripe.key,
	}

	// get payment intent
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// get payment method
	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// get card details
	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// create transaction data
	txnData = TransactionData{
		FirstName: firstName,
		LastName: lastName,
		Email: email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount: amount,
		PaymentCurrency: paymentCurrency,
		LastFour: lastFour,
		ExpiryMonth: int(expiryMonth),
		ExpiryYear: int(expiryYear),
		BankReturnCode: pi.Charges.Data[0].ID,
	}

	return txnData, nil
}

// PaymentSucceeded displays the recipet page after purchasing a widget
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read submitted form data
	widgetID, err := strconv.Atoi(r.Form.Get("product-id"))
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new transaction
	txn := models.Transaction{
		Amount: txnData.PaymentAmount,
		Currency: txnData.PaymentCurrency,
		LastFour: txnData.LastFour,
		ExpiryMonth: txnData.ExpiryMonth,
		ExpiryYear: txnData.ExpiryYear,
		BankReturnCode: txnData.BankReturnCode,
		PaymentIntent: txnData.PaymentIntentID,
		PaymentMethod: txnData.PaymentMethodID,
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
		Quantity: 1, // TODO: can only buy one for now
		Amount: txnData.PaymentAmount,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// write receipt data to session
	app.Session.Put(r.Context(), "receipt", txnData)
	// redirect user to new page
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

// VirtualTerminalPaymentSucceeded displays the recipet page for virtual terminal txns
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create new transaction
	txn := models.Transaction{
		Amount: txnData.PaymentAmount,
		Currency: txnData.PaymentCurrency,
		LastFour: txnData.LastFour,
		ExpiryMonth: txnData.ExpiryMonth,
		ExpiryYear: txnData.ExpiryYear,
		BankReturnCode: txnData.BankReturnCode,
		PaymentIntent: txnData.PaymentIntentID,
		PaymentMethod: txnData.PaymentMethodID,
		TransactionStatusID: 2, // transaction_status = "cleared"
	}
	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// write receipt data to session
	app.Session.Put(r.Context(), "receipt", txnData)
	// redirect user to new page
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

// Receipt renders the receipt page after credit card purchase
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	// get transaction data from session and to template map
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminalReceipt renders the VirtualTerminalReceipt page after credit card purchase
func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {
	// get transaction data from session and to template map
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{
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

// BronzePlan renders the bronze plan purchase page
func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DB.GetWidget(2)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// BronzePlanReceipt renders the bronze plan recipet page
func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// LoginPage displays the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}