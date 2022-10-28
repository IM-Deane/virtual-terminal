package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
	}
	stripe struct {
		secret string
		key string
	}
}

type application struct {
	config config
	infoLog *log.Logger
	errorLog *log.Logger
	version string
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		IdleTimeout: 30 * time.Second,
		ReadTimeout: 10 * time.Second,
		ReadHeaderTimeout: 5 *time.Second,
		WriteTimeout: 5 * time.Second,
	}

	app.infoLog.Printf("Starting Backend server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()
}


func main() {
	var cfg config

	// load env file
	envErr := godotenv.Load(".env")
	if envErr != nil {
		log.Fatalf("Error occured while loading .env file! Err: %s", envErr)
	}

	// setup
	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application env {development | production|maintenance}")

	flag.Parse()

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	// logger
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// set app config
	app := &application{
		config: cfg,
		infoLog: infoLog,
		errorLog: errorLog,
		version: version,
	}

	err := app.serve()
	if err != nil {
		log.Fatal(err)
	}
}