package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.octaviorassi.net/internal/models"
	"snippetbox.octaviorassi.net/internal/validator"
)

// The struct's fields must be exported in order to be read by the html/template package
type snippetCreateForm struct {
	Title		string
	Content		string
	Expires		int
	validator.Validator
}

// After implementing the application struct, instead of writing functions as standalone functions, we
// define them as the methods of the application class. Note that struct is not an interface, when we
// defined it we did not specify which methods it should implement, instead we can directly implement them.

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
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

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, r, http.StatusOK, "view.tmpl.html", data)

}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {

	/* 	We must pass an initialized templateData with a non-nil Form in order to have
	the template correctly render the first time. We set a default 365 expire time	*/

	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{ Expires: 365, }

	app.render(w, r, http.StatusOK, "create.tmpl.html", data)
	
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	title 	:= r.PostForm.Get("title")
	content := r.PostForm.Get("content")

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Initializes a map for error storage
	form := snippetCreateForm{
		Title: 		 title,
		Content: 	 content,
		Expires: 	 expires,
	}

	// Validate the fields
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must be equal to 1, 7, or 365")

	// Check for any errors. If there are any, re-render the template highlighting them
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	// Else, insert the snippet and redirect the user
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)

}
