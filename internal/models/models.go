package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *pgx.Conn
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModels(db *pgx.Conn) Models {
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
type Transactions struct {
	ID int `json:"id"`
	Amount int `json:"amount"`
	Currency string `json:"currency"`
	LastFour string `json:"last_four"`
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

func (m *DBModel) GetWidget(id int) (Widget, error) {
	// timeout after 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRow(ctx, "select id, name from widgets where id=$1", id)
	err := row.Scan(&widget.ID, &widget.Name)
	if err != nil {
		return widget, err
	}

	return widget, nil
}