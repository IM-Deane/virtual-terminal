package cards

import (
	"github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/paymentintent"
	"github.com/stripe/stripe-go/v73/paymentmethod"
)

type Card struct {
	Secret string
	Key string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount int
	Currency string
	LastFour string
	BankReturnCode string
}

// Charge alias for CreatePaymentIntent with a more meaningful name
func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

// CreatePaymentIntent handles stripe payment
func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	// create payment intent
	params := &stripe.PaymentIntentParams{
		Amount: stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	// NOTE: can add additional props by: params.AddMetadata("key", "value")

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeError, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeError.Code)
		}
		return nil, msg, err
	}
	return pi, "", nil
}

// GetPaymentMethod gets the payment method by payment intent id
func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	// set stripe key
	stripe.Key = c.Secret

	pm, err := paymentmethod.Get(s, nil)
	if err != nil {
		return nil, err
	}
	return pm, nil
}

// RetrievePaymentIntent gets an *existing* payment intent by id
func (c *Card) RetrievePaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// cardErrorMessage helper for rendering stripe error messages
func cardErrorMessage(code stripe.ErrorCode) string {
	var msg =""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired"
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Incorrect CVC code"
	case stripe.ErrorCodeIncorrectZip:
		msg = "Incorrect zip/postal code"
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to charge to your card"
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to charge to your card"
	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Insufficient balance"
	case stripe.ErrorCodePostalCodeInvalid:
		msg = "Your postal code is invalid"
	default:
		msg = "Your card was declined"
	}
	return msg
}