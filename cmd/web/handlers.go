package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"snippetbox.octaviorassi.net/internal/models"
)

// After implementing the application struct, instead of writing functions as standalone functions, we
// define them as the methods of the application class. Note that struct is not an interface, when we
// defined it we did not specify which methods it should implement, instead we can directly implement them.

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	for _, snippet := range snippets {
		fmt.Fprintf(w, "%+v\n", snippet)
	}


	/*
	// Create a slice with the paths to the html files
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
	}

	// Using files... rather than files makes the contents of the slice be passed as variadic arguments (separated)
	// ParseFiles takes an arbitrary amount of file paths and returns a *Template object, which we store in ts.
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// ExecuteTemplate is a method for the Template object which executes the template and writes it into the
	// io.Writer object passed as its first argument. In this case, it is executing the template (replacing the
	// placeholders with the actual contents) and then writing it as a response to the http request through w.
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, r, err)
	}
	*/
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
		// Extract the id from the path's wildcard value
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	// Query the DB for the ID and check for possible errors
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// Check if no rows were found
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/view.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{ Snippet: snippet, }
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r ,err)
	}

}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a form for creating a new snippet..."))
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Example with dummy variables which we will later replace
	title 	:= "O snail"
	content := "O snail\nClimb Mount FUji,\nBut slowly, slowly!\n\n- Kobayashi Issa"
	expires := 7

	// We now call the snippet.Insert from the app.DB 
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect the user to the relevant page for the snippet
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
