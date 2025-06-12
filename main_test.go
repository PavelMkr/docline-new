//go:build !ui
// +build !ui

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_Endpoints(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)
	endpoints := []string{"/upload", "/heuristic", "/ngram", "/automatic", "/interactive"}
	for _, endpoint := range endpoints {
		req := httptest.NewRequest("POST", endpoint, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Endpoint %s returned status %d", endpoint, resp.StatusCode)
		}
	}
}
