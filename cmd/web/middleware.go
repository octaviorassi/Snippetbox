package main

import (
	"fmt"
	"log/slog"
	"net/http"
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