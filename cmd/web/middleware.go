package main

import (
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
