package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"snippetbox.octaviorassi.net/internal/models"
	"snippetbox.octaviorassi.net/internal/validator"
)

const MinPassLength = 8

// The struct's fields must be exported in order to be read by the html/template package
type snippetCreateForm struct {
	Title		string	`form:"title"`
	Content		string	`form:"content"`
	Expires		int		`form:"expires"`
	validator.Validator	`form:"-"`
}

type userSignUpForm struct {
	Name 		string	`form:"name"`
	Email 		string	`form:"email"`
	Password 	string 	`form:"password"`
	validator.Validator	`form:"-"`
}

type userLoginForm struct {
	Email 		string 	`form:"email"`
	Password 	string 	`form:"password"`
	validator.Validator	`form:"-"`
}


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

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignUpForm{}
	app.render(w, r, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {

	var form userSignUpForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	form.CheckField(validator.Matches(form.Email, validator.EmailRx), "email", "This field must be a valid email address")
	form.CheckField(validator.MinChars(form.Password, MinPassLength), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	// Load the user into the database (placeholder)
	id, err := app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {

		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address already in use")

			data := app.newTemplateData(r) 
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)

		} else {
			app.serverError(w, r, err)
		}

	}

	// If the user was successfully signed up, log it and generate a flash message notifying them
	app.logger.Info("loaded user:", slog.Any("id", id), slog.Any("email", form.Email), slog.Any("name", form.Name))

	app.sessionManager.Put(r.Context(), "flash", "Your signup was sucessfull. Please, log in.")

	// And redirect them to log in
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}


func (app *application) userLogIn(w http.ResponseWriter, r *http.Request) {
	/* renders the page and the form */
	data 	  := app.newTemplateData(r)
	data.Form  = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) userLogInPost(w http.ResponseWriter, r *http.Request) {
	/* parses the form and checks the database */
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Check for minimum password and email requirements
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRx), "email", "This field must be a valid email address")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	// If the fields are well formatted, try to authenticate the user
	id, err := app.users.Authenticate(form.Email, form.Password)

	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	// Renew the session id since the user privilege level changed
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r , err)
		return
	}

	// And add the new id to the user's session
	app.sessionManager.Put(r.Context(), "authenticatedUserId", id)

	// Finally, redirect the user to the snippet creation page
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)

}

func (app *application) userLogOutPost(w http.ResponseWriter, r *http.Request) {
	// Renew the session ID first
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}


	// We only have to remove the user's authenticatedUserId header
	app.sessionManager.Remove(r.Context(), "authenticatedUserId")

	// Notify them through a flash message
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out sucessfully")

	// And redirect them to the landing page
	http.Redirect(w, r, "/", http.StatusSeeOther)

}


func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {

	/* 	We must pass an initialized templateData with a non-nil Form in order to have
	the template correctly render the first time. We set a default 365 expire time	*/

	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{ Expires: 365, }

	app.render(w, r, http.StatusOK, "create.tmpl.html", data)
	
}


func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// The decode method fills the form fields with their corresponding values from the HTML form
	var form snippetCreateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	
	// Validate the fields
	form.CheckField(validator.NotBlank(form.Title), "title",
					"This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title",
					"This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content",
					"This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires",
					"This field must be equal to 1, 7, or 365")

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

	// Add the flash message to the session data
	app.sessionManager.Put(r.Context(), "flash", "Snippet sucessfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)

}
