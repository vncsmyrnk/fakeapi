package middleware

import (
	"fakeapi/internal/version"
	"net/http"
)

const versionHeaderKey = "x-fakeapi-version"

func VersionResponseHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(versionHeaderKey, version.ServerVersion)
		next.ServeHTTP(w, r)
	})
}
