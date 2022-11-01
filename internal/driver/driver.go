package driver

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func OpenDB(dsn string) (*pgx.Conn, error) {
	db, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}