package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/kientink26/go-json-api/cmd/api/application"
	"github.com/kientink26/go-json-api/cmd/api/config"
	"github.com/kientink26/go-json-api/internal/data"
	"github.com/kientink26/go-json-api/internal/mailer"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	// Declare an instance of the Config struct.
	var cfg config.Config
	flag.IntVar(&cfg.Port, "port", 0, "API server port")
	flag.StringVar(&cfg.Env, "env", "", "Environment (development|staging|production)")
	flag.StringVar(&cfg.Db.Dsn, "db-dsn", "", "PostgreSQL DSN")
	// Read the SMTP server configuration settings into the Config struct, using the
	// Mailtrap settings as the default values
	flag.StringVar(&cfg.Smtp.Host, "smtp-host", "", "SMTP host")
	flag.IntVar(&cfg.Smtp.Port, "smtp-port", 0, "SMTP port")
	flag.StringVar(&cfg.Smtp.Username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.Smtp.Password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.Smtp.Sender, "smtp-sender", "", "SMTP sender")
	flag.StringVar(&cfg.Cors.TrustedOrigin, "cors-trusted-origin", "", "Trusted CORS origin")
	flag.Parse()

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exits.
	defer db.Close()
	logger.Printf("database connection pool established")

	app := &application.Application{
		Config: cfg,
		Logger: logger,
		Models: data.NewModels(db),
		Mailer: mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("starting %s server on %s", cfg.Env, srv.Addr)
	err = srv.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}

// The openDB() function returns a sql.DB connection pool.
func openDB(cfg config.Config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config
	// struct.
	db, err := sql.Open("postgres", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}
	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	// Return the sql.DB connection pool.
	return db, nil
}
