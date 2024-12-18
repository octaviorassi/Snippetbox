package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/form/v4"
)

func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Parse the form
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Decode it into the destination
	err = app.formDecoder.Decode(dst, r.PostForm)

	// Check for the possible InvalidDecoderError
	if err != nil {
		// If the error generated is an InvalidDecoderError, panic
		var InvalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &InvalidDecoderError) {
			panic(err)
		}

		// Else, just return it
		return err
	}

	return nil
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	// Write the template into the buffer to check for possible errors first
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If the template loaded correctly, write the buffer's contents to w
	w.WriteHeader(status)
	buf.WriteTo(w)
	
}

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