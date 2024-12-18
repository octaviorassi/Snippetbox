package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"

	"snippetbox.octaviorassi.net/internal/models"
)

type application struct {
	logger 			*slog.Logger
	snippets 		*models.SnippetModel
	templateCache 	templateCache
	formDecoder		*form.Decoder
	sessionManager  *scs.SessionManager
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

	// Create the snippetModel based on db
	snippetModel, err := models.NewSnippetModel(db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Defer the closure of all the prepared statements
	defer snippetModel.InsertStmt.Close()
	defer snippetModel.GetStmt.Close()
	defer snippetModel.LatestStmt.Close()

	// Start the template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize a decoder instance
	formDecoder := form.NewDecoder()

	// Initialize a new session manager and set it up to work on our MySQL db
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:	  		logger,
		snippets: 		snippetModel,
		templateCache: 	templateCache,
		formDecoder: 	formDecoder,
		sessionManager: sessionManager,
	}


	mux := app.routes()

	// Initialize a tlsConfig for non-default TLS settings
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	
	// Initialize a new http.Server struct
	srv := &http.Server{
		Addr: 	 *addr,
		Handler: mux,
		// ErrorLog acts as a bridge between the old logger used by http and our applications
		// new structured logger. The http server logs are now written to our logger at Error level.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	logger.Info("Starting server", "addr", *addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

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