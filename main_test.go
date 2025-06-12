//go:build !ui
// +build !ui

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestIntegration_FileUpload(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	// create a temporary .xml file with valid content
	tmpFile, err := os.CreateTemp("", "test_*.xml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := `<?xml version="1.0" encoding="UTF-8"?><root><item>test</item></root>`
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tmpFile.Close()

	testCases := []struct {
		name           string
		endpoint       string
		contentType    string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Valid JSON upload to heuristic",
			endpoint:       "/heuristic",
			contentType:    "application/json",
			requestBody:    fmt.Sprintf(`{"extension_point_checkbox": true, "file_path": "%s"}`, tmpFile.Name()),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON upload to heuristic",
			endpoint:       "/heuristic",
			contentType:    "application/json",
			requestBody:    `{"invalid": "json"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty JSON upload",
			endpoint:       "/heuristic",
			contentType:    "application/json",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.requestBody)
			req := httptest.NewRequest("POST", tc.endpoint, body)
			req.Header.Set("Content-Type", tc.contentType)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			resp := w.Result()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d for test case %s", tc.expectedStatus, resp.StatusCode, tc.name)
			}
		})
	}
}

func TestIntegration_HTTPMethods(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	endpoints := []string{"/upload", "/heuristic", "/ngram", "/automatic", "/interactive"}
	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, endpoint := range endpoints {
		for _, method := range methods {
			t.Run(fmt.Sprintf("%s %s", method, endpoint), func(t *testing.T) {
				req := httptest.NewRequest(method, endpoint, nil)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)
				resp := w.Result()

				// All endpoints should only accept POST
				if resp.StatusCode != http.StatusMethodNotAllowed {
					t.Errorf("Expected MethodNotAllowed for %s %s, got %d",
						method, endpoint, resp.StatusCode)
				}
			})
		}
	}
}

func TestIntegration_ResponseContentType(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_*.xml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some valid XML content
	testContent := `<?xml version="1.0" encoding="UTF-8"?><root><item>test</item></root>`
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tmpFile.Close()

	// Test cases for successful responses
	testCases := []struct {
		name         string
		endpoint     string
		contentType  string
		requestBody  string
		expectedType string
	}{
		{
			name:         "Heuristic endpoint content type",
			endpoint:     "/heuristic",
			contentType:  "application/json",
			requestBody:  fmt.Sprintf(`{"extension_point_checkbox": true, "file_path": "%s"}`, tmpFile.Name()),
			expectedType: "application/json",
		},
		{
			name:         "Ngram endpoint content type",
			endpoint:     "/ngram",
			contentType:  "application/json",
			requestBody:  fmt.Sprintf(`{"min_clone_slider": 2, "max_edit_slider": 1, "max_fuzzy_slider": 1, "source_language": "english", "file_path": "%s"}`, tmpFile.Name()),
			expectedType: "application/json",
		},
		{
			name:         "Automatic endpoint content type",
			endpoint:     "/automatic",
			contentType:  "application/json",
			requestBody:  fmt.Sprintf(`{"min_clone_length": 2, "archetype_length": 1, "strict_filter": true, "file_path": "%s"}`, tmpFile.Name()),
			expectedType: "application/json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.requestBody)
			req := httptest.NewRequest("POST", tc.endpoint, body)
			req.Header.Set("Content-Type", tc.contentType)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			resp := w.Result()

			// Only check content type for successful responses
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.expectedType) {
					t.Errorf("Expected content type %s, got %s for endpoint %s",
						tc.expectedType, contentType, tc.endpoint)
				}
			}
		})
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	testCases := []struct {
		name           string
		endpoint       string
		requestBody    string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "Missing content type",
			endpoint:       "/heuristic",
			requestBody:    "test content",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid content type",
			endpoint:       "/heuristic",
			requestBody:    "test content",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent endpoint",
			endpoint:       "/nonexistent",
			requestBody:    "",
			contentType:    "application/xml",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.NewReader(tc.requestBody)
			req := httptest.NewRequest("POST", tc.endpoint, body)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			resp := w.Result()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d for test case %s",
					tc.expectedStatus, resp.StatusCode, tc.name)
			}
		})
	}
}
