package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncUpdate(t *testing.T) {
	// Test that IncUpdate doesn't panic
	assert.NotPanics(t, func() {
		IncUpdate()
	})
}

func TestIncTransactionsSaved(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{
			name:   "success status",
			status: "success",
		},
		{
			name:   "error status",
			status: "error",
		},
		{
			name:   "empty status",
			status: "",
		},
		{
			name:   "special characters status",
			status: "status_with_special_chars_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that IncTransactionsSaved doesn't panic
			assert.NotPanics(t, func() {
				IncTransactionsSaved(tt.status)
			})
		})
	}
}

func TestHandler(t *testing.T) {
	// Test that Handler returns a valid HTTP handler
	handler := Handler()
	assert.NotNil(t, handler)

	// Test that the handler can handle requests
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	// Should return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Should return some content (metrics data)
	assert.NotEmpty(t, w.Body.String())
}

func TestHandler_WithDifferentMethods(t *testing.T) {
	handler := Handler()
	
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/metrics", nil)
			w := httptest.NewRecorder()
			
			handler.ServeHTTP(w, req)
			
			// Prometheus handler should handle all methods appropriately
			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to metrics functions
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			IncUpdate()
			IncTransactionsSaved("concurrent_test")
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should not panic
	assert.True(t, true)
}

func TestMetrics_Integration(t *testing.T) {
	// Test integration of all metrics functions
	assert.NotPanics(t, func() {
		// Increment various metrics
		IncUpdate()
		IncUpdate()
		IncTransactionsSaved("success")
		IncTransactionsSaved("error")
		IncTransactionsSaved("pending")
		
		// Get handler and make request
		handler := Handler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		
		handler.ServeHTTP(w, req)
		
		// Should return 200 OK
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
