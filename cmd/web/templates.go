package main

import (
	"path/filepath"
	"text/template"

	"snippetbox.octaviorassi.net/internal/models"
)

type TemplateCache = map[string]*template.Template

/*	newTemplateCache initializes the in-memory template cache, returning a map with
	the page names as keys and their associated Template sets as values	*/
func newTemplateCache() (TemplateCache, error) {
	// Initialize an empty map that will act as a cache
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		
		name := filepath.Base(page)

		// Parse the base template to create a template set*
		ts, err := template.ParseFiles("./ui/html/base.tmpl.html")
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

type templateData struct {
	Snippet	   models.Snippet
	Snippets []models.Snippet
}