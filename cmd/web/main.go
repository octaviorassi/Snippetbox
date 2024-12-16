package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger *slog.Logger
}

func main() {
	// Define and parse the execution flags
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn	 := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()
	
	// Create the app's logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ AddSource: true,}))

	// Create the DB connection pool
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Closes the database when the surrounding function (i.e. main) finishes its execution
	defer db.Close()

	// Create the app and load the handlers into the mux
	app := &application{ logger: logger, }
	mux := app.routes()

	// Start the server at the provided address with the defined handlers
	logger.Info("Starting server", "addr", *addr)
	err = http.ListenAndServe(*addr, mux)

	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	// Ping to test and establish a connection, since they are established lazily
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}