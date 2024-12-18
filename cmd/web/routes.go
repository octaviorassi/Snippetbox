package main

import (
	"net/http"

	"github.com/justinas/alice"
) 

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	
	// Define a chain of middleware standard for all requests and apply it to mux
	standard := alice.New(app.panicRecover, app.logRequest, commonHeaders)
	
	// And a chain of middleware standard for all dynamic requests, i.e. not those fetching on static
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Apply 'dynamic' to all non-static routes
	mux.Handle("GET /{$}", 				 dynamic.ThenFunc(app.home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	mux.Handle("GET /snippet/create", 	 dynamic.ThenFunc(app.snippetCreate))
	mux.Handle("POST /snippet/create", 	 dynamic.ThenFunc(app.snippetCreatePost))

	return standard.Then(mux)
}