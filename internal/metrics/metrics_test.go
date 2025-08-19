package metrics

import (
	"net/http/httptest"
	"testing"
)

func TestMetrics_HandlerAndCounters(t *testing.T) {
	// Exercise counters
	IncUpdate()
	IncTransactionsSaved("ok")

	rec := httptest.NewRecorder()
	h := Handler()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != 200 {
		t.Fatalf("metrics handler status: %d", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatalf("expected metrics body")
	}
}
