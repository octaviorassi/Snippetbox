package main

import (
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/justinas/nosurf"
	"snippetbox.octaviorassi.net/internal/models"
)


type templateData struct {
	Snippet	   		models.Snippet
	Snippets 		[]models.Snippet
	CurrentYear 	int
	Form 			any
	Flash			string
	IsAuthenticated bool
	CSRFToken		string
}

type templateCache = map[string]*template.Template

/*	newTemplateData returns an initialized templateData object with CurrentYear set
	to the current year and Flash set to the request's flash value, if it exists	*/
func (app *application) newTemplateData(r *http.Request) templateData {
	return	templateData{
				CurrentYear: 	 time.Now().Year(),
				Flash:		  	 app.sessionManager.PopString(r.Context(), "flash"),
				IsAuthenticated: app.isAuthenticated(r),
				CSRFToken: 		 nosurf.Token(r),	
			}
}

/*	newTemplateCache initializes the in-memory template cache, returning a map with
	the page names as keys and their associated Template sets as values	*/
func newTemplateCache() (templateCache, error) {
	// Initialize an empty map that will act as a cache
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		
		name := filepath.Base(page)

		/* Load our template functions into a template set first and then parse the base template*/
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}

		// Call ParseGlob on this template set to add any partials
		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl.html")
		if err != nil {
			return nil, err
		}

		// Now parse the corresponding page template
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

/*	Defines a format string for the representation of time.Time objects */
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

/*	Define a global map that matches strings to our template functions	*/
var functions = template.FuncMap{ "humanDate": humanDate, }