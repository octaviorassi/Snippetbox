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
	mux.Handle("GET /user/signup", 		 dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", 	 dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", 		 dynamic.ThenFunc(app.userLogIn))
	mux.Handle("POST /user/login", 		 dynamic.ThenFunc(app.userLogInPost))
	mux.Handle("POST /user/logout",		 dynamic.ThenFunc(app.userLogOutPost))



	return standard.Then(mux)
}