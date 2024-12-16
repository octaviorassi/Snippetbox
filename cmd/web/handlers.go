package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// After implementing the application struct, instead of writing functions as standalone functions, we
// define them as the methods of the application class. Note that struct is not an interface, when we
// defined it we did not specify which methods it should implement, instead we can directly implement them.

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

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

}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	// Aprovecho que w satisface la intefaz io.Writer para pasarlo como argumento a Fprint
	fmt.Fprint(w, "Display a specific snippet with ID %d...", id)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a form for creating a new snippet..."))
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Save a new snippet..."))
}
