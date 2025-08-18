package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	updatesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "bot_updates_total",
			Help: "Total number of Telegram updates processed",
		},
	)

	transactionsSaved = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_transactions_saved_total",
			Help: "Total number of transactions saved",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(updatesTotal)
	prometheus.MustRegister(transactionsSaved)
}

// IncUpdate increments updates counter.
func IncUpdate() { updatesTotal.Inc() }

// IncTransactionsSaved increments saved counter with a status label.
func IncTransactionsSaved(status string) { transactionsSaved.WithLabelValues(status).Inc() }

// Handler returns the HTTP handler for /metrics.
func Handler() http.Handler { return promhttp.Handler() }


