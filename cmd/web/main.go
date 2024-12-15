package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ AddSource: true,}))

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /snippet/view/{id}", snippetView)
	mux.HandleFunc("GET /snippet/create", snippetCreate)


	// given an http.Handler, StripPrefix generates another one that serves the same purpose but removes the 
	// given prefix from the request's path.
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("POST /snippet/create", snippetCreatePost)

	logger.Info("Starting server", "addr", *addr)
	err := http.ListenAndServe(*addr, mux)

	// there is no equivalent to the log.Fatal() using slog, but we can log the error at the Error severity level
	// and then prompt the system to exit with an abnormal termination status (1).
	logger.Error(err.Error())
	os.Exit(1)
}