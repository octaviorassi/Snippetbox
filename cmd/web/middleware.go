package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/justinas/nosurf"
)

const (
    // Header names
    HeaderContentSecurityPolicy = "Content-Security-Policy"
    HeaderReferrerPolicy        = "Referrer-Policy"
    HeaderXContentTypeOptions   = "X-Content-Type-Options"
    HeaderXFrameOptions         = "X-Frame-Options"
    HeaderXXSSProtection        = "X-XSS-Protection"
	HeaderServer				= "Server"

    // Policy values
    ContentSecurityPolicyValue = "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
    ReferrerPolicyValue        = "origin-when-cross-origin"
    XContentTypeOptionsValue   = "nosniff"
    XFrameOptionsValue         = "deny"
    XXSSProtectionValue        = "0"
	Go						   = "Go"
)


func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// If there is no id to begin with, simply continue with the next handler
				id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
				if id == 0 {
					next.ServeHTTP(w, r)
					return
				}

				// Otherwise, check if that id actually exists in the database
				exists, err := app.users.Exists(id)
				if err != nil {
					app.serverError(w, r, err)
					return
				}

				// If the id belongs to a user, add that to the context and continue with next
				if exists {
					ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
					r = r.WithContext(ctx)
				}

				// Otherwise, ignore it and simply continue
				next.ServeHTTP(w, r)
			})
}

/*	noSurf includes a customized CSRF cookie to the given handler to prevent CSRF attacks	*/
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: true,
	})

	return csrfHandler
}

/*	requireAuthentication adds an authentication check to the given handler. If it passes, the
	next handler executes. Otherwise, the user is redirected to the login page instead.	*/
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// If the user is not authenticated, then redirect them to login
				if !app.isAuthenticated(r) {
					http.Redirect(w, r, "/user/login", http.StatusSeeOther)
					return
				}

				// If they are, for security reasons set caching to false
				w.Header().Add("Cache-Control", "no-store")

				// And continue with the next handler
				next.ServeHTTP(w, r)
			})
}

/*	commonHeaders adds several basic headers to the given handler, including security headers.	*/
func commonHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(HeaderContentSecurityPolicy, ContentSecurityPolicyValue)
				w.Header().Set(HeaderReferrerPolicy, ReferrerPolicyValue)
				w.Header().Set(HeaderXContentTypeOptions, XContentTypeOptionsValue)
				w.Header().Set(HeaderXFrameOptions, XFrameOptionsValue)
				w.Header().Set(HeaderXXSSProtection, XXSSProtectionValue)
				w.Header().Set(HeaderServer, Go)

				next.ServeHTTP(w, r)
			})
}

/*	logRequest modifies the given handler to make it log the request they receive	*/
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				app.logger.Info("Request received", slog.Any("method", r.Method),
													slog.Any("ip", r.RemoteAddr),
													slog.Any("proto", r.Proto),
													slog.Any("uri", r.URL.RequestURI()))

				next.ServeHTTP(w, r)
			})
}

/*	panicRecover modifies the handler given to close the connection and log the error in
	case of a panic() call	*/
func (app *application) panicRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if err := recover(); err != nil {
						// Trigger Go's HTTP server to close this connection and inform the user
						w.Header().Set("Connection", "close")

						// Generate a proper Interval Server Error response
						app.serverError(w, r, fmt.Errorf("%s", err))
					}
				}()

				next.ServeHTTP(w, r)
			})
}