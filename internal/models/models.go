package models

import (
	"context"
	"database/sql"
	"time"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}


// Widget is the scheme for all widgets
type Widget struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	InventoryLevel int `json:"inventory_level"`
	Price int `json:"price"`
	Image string `json:"image"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Order is the type for all orders
type Order struct {
	ID int `json:"id"`
	WidgetID int `json:"widget_id"`
	TransactionID int `json:"transaction_id"`
	CustomerID int `json:"customer_id"`
	StatusID int `json:"status_id"`
	Quantity int `json:"quantity"`
	Amount int `json:"amount"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Status is the type for statuses
type Status struct {
	ID int `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TransactionStatus is the type for transaction statuses
type TransactionStatus struct {
	ID int `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transactions is the type for transactions
type Transaction struct {
	ID int `json:"id"`
	Amount int `json:"amount"`
	Currency string `json:"currency"`
	LastFour string `json:"last_four"`
	ExpiryMonth int `json:"expiry_month"`
	ExpiryYear int `json:"expiry_year"`
	BankReturnCode string `json:"bank_return_code"`
	TransactionStatusID int `json:"transaction_status_id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID int `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Password string `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is the type for customers
type Customer struct {
	ID int `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	// timeout after 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `
	SELECT
		id, name, description, inventory_level, price, COALESCE(image, ''),
		created_at, updated_at
	FROM
		widgets
	WHERE id = $1`, id)
	err := row.Scan(
		&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Price,
		&widget.Image,
		&widget.CreatedAt,
		&widget.UpdatedAt,
	)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

// InsertTransaction inserts a new txn into db and returns the txn id
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	// timeout after 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	stmt := `
		INSERT INTO transactions
			(amount, currency, last_four, bank_return_code,
			transaction_status_id, expiry_month, expiry_year,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	id := 0
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.TransactionStatusID,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		time.Now(), // created_at
		time.Now(), // updated_at
	).Scan(&id)

	if err != nil {
		return id, err
	}

	return int(id), nil
}


// InsertOrder inserts a new order into db and returns the order id
func (m *DBModel) InsertOrder(order Order) (int, error) {
	// timeout after 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	stmt := `
		INSERT INTO orders
			(widget_id, transaction_id, status_id, quantity,
			amount, customer_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	id := 0
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		order.CustomerID,
		time.Now(), // created_at
		time.Now(), // updated_at
	).Scan(&id)
	if err != nil {
		return id, err
	}

	return int(id), nil
}


// InsertCustomer inserts a new customer into db and returns customer id
func (m *DBModel) InsertCustomer(c Customer) (int, error) {
	// timeout after 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	stmt := `
		INSERT INTO customers
			(first_name, last_name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	id := 0
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		c.FirstName,
		c.LastName,
		c.Email,
		time.Now(), // created_at
		time.Now(), // updated_at
	).Scan(&id)
	if err != nil {
		return id, err
	}

	return int(id), nil
}