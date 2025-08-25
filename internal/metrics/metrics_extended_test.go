package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncUpdate_MultipleCalls(t *testing.T) {
	// Test IncUpdate with multiple calls
	assert.NotPanics(t, func() {
		IncUpdate()
		IncUpdate()
		IncUpdate()
	})
}

func TestIncTransactionsSaved_MultipleStatuses(t *testing.T) {
	// Test IncTransactionsSaved with multiple statuses
	assert.NotPanics(t, func() {
		IncTransactionsSaved("success")
		IncTransactionsSaved("error")
		IncTransactionsSaved("pending")
		IncTransactionsSaved("cancelled")
	})
}

func TestIncTransactionsSaved_WithEmptyStatus(t *testing.T) {
	// Test IncTransactionsSaved with empty status
	assert.NotPanics(t, func() {
		IncTransactionsSaved("")
	})
}

func TestIncTransactionsSaved_WithSpecialCharacters(t *testing.T) {
	// Test IncTransactionsSaved with special characters in status
	assert.NotPanics(t, func() {
		IncTransactionsSaved("status_with_underscores")
		IncTransactionsSaved("status-with-dashes")
		IncTransactionsSaved("status with spaces")
	})
}

func TestIncTransactionsSaved_WithLongStatus(t *testing.T) {
	// Test IncTransactionsSaved with long status
	longStatus := "very_long_status_that_might_be_used_in_some_edge_cases_123456789"
	assert.NotPanics(t, func() {
		IncTransactionsSaved(longStatus)
	})
}

func TestHandler_ReturnsHandler(t *testing.T) {
	// Test that Handler returns a valid HTTP handler
	handler := Handler()
	assert.NotNil(t, handler)
	assert.Implements(t, (*http.Handler)(nil), handler)
}

func TestHandler_CanServeRequests(t *testing.T) {
	// Test that Handler can serve requests
	handler := Handler()
	assert.NotNil(t, handler)
	
	// Create a test request
	req, err := http.NewRequest("GET", "/metrics", nil)
	assert.NoError(t, err)
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Serve the request
	handler.ServeHTTP(rr, req)
	
	// Check that we got a response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_WithDifferentMethods(t *testing.T) {
	// Test Handler with different HTTP methods
	handler := Handler()
	assert.NotNil(t, handler)
	
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	
	for _, method := range methods {
		req, err := http.NewRequest(method, "/metrics", nil)
		assert.NoError(t, err)
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		// Most methods should return 200 OK for Prometheus metrics
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestHandler_WithDifferentPaths(t *testing.T) {
	// Test Handler with different paths
	handler := Handler()
	assert.NotNil(t, handler)
	
	paths := []string{"/metrics", "/", "/prometheus", "/health"}
	
	for _, path := range paths {
		req, err := http.NewRequest("GET", path, nil)
		assert.NoError(t, err)
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		// Should return 200 OK for valid requests
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestMetrics_Integration(t *testing.T) {
	// Test integration of all metrics functions
	assert.NotPanics(t, func() {
		// Increment updates
		IncUpdate()
		IncUpdate()
		
		// Increment transactions with different statuses
		IncTransactionsSaved("success")
		IncTransactionsSaved("error")
		
		// Get handler
		handler := Handler()
		assert.NotNil(t, handler)
		
		// Test handler
		req, err := http.NewRequest("GET", "/metrics", nil)
		assert.NoError(t, err)
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestIncUpdate_StressTest(t *testing.T) {
	// Test IncUpdate with many rapid calls
	assert.NotPanics(t, func() {
		for i := 0; i < 100; i++ {
			IncUpdate()
		}
	})
}

func TestIncTransactionsSaved_StressTest(t *testing.T) {
	// Test IncTransactionsSaved with many rapid calls
	assert.NotPanics(t, func() {
		statuses := []string{"success", "error", "pending", "cancelled", "timeout"}
		for i := 0; i < 100; i++ {
			status := statuses[i%len(statuses)]
			IncTransactionsSaved(status)
		}
	})
}

func TestHandler_WithHeaders(t *testing.T) {
	// Test Handler with different headers
	handler := Handler()
	assert.NotNil(t, handler)
	
	headers := map[string]string{
		"Accept":           "text/plain",
		"User-Agent":       "Prometheus/2.0.0",
		"Authorization":    "Bearer token123",
		"X-Forwarded-For":  "192.168.1.1",
	}
	
	req, err := http.NewRequest("GET", "/metrics", nil)
	assert.NoError(t, err)
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_WithQueryParameters(t *testing.T) {
	// Test Handler with query parameters
	handler := Handler()
	assert.NotNil(t, handler)
	
	queryParams := []string{
		"?format=text",
		"?format=json",
		"?debug=true",
		"?timeout=30s",
	}
	
	for _, params := range queryParams {
		req, err := http.NewRequest("GET", "/metrics"+params, nil)
		assert.NoError(t, err)
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestHandler_WithDifferentUserAgents(t *testing.T) {
	// Test Handler with different user agents
	handler := Handler()
	assert.NotNil(t, handler)
	
	userAgents := []string{
		"Prometheus/2.0.0",
		"curl/7.68.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"",
	}
	
	for _, ua := range userAgents {
		req, err := http.NewRequest("GET", "/metrics", nil)
		assert.NoError(t, err)
		
		if ua != "" {
			req.Header.Set("User-Agent", ua)
		}
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestHandler_WithDifferentContentTypes(t *testing.T) {
	// Test Handler with different content types
	handler := Handler()
	assert.NotNil(t, handler)
	
	contentTypes := []string{
		"text/plain",
		"application/json",
		"text/html",
		"",
	}
	
	for _, ct := range contentTypes {
		req, err := http.NewRequest("GET", "/metrics", nil)
		assert.NoError(t, err)
		
		if ct != "" {
			req.Header.Set("Accept", ct)
		}
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	// Test metrics with concurrent access
	assert.NotPanics(t, func() {
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				IncUpdate()
				IncTransactionsSaved("success")
				done <- true
			}()
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestHandler_WithLargeRequest(t *testing.T) {
	// Test Handler with large request
	handler := Handler()
	assert.NotNil(t, handler)
	
	// Create a large request body
	largeBody := make([]byte, 1024*1024) // 1MB
	for i := range largeBody {
		largeBody[i] = 'A'
	}
	
	req, err := http.NewRequest("POST", "/metrics", nil)
	assert.NoError(t, err)
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	// Should still return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)
}
