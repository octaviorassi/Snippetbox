package main

import (
	"log/slog"
	"net/http"
)

/*	serverError writes a log entry at Error level describing the request's method and URI
	and responds to the request with a generic 500 Internal Server Error to the user	  */
func (app *application) serverError (w http.ResponseWriter, r *http.Request, err error) {

	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		// trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), slog.Any("method", method), slog.Any("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

/*	clientError sends a specific status code along with its corresponding description to the user */
func (app *application) clientError (w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}