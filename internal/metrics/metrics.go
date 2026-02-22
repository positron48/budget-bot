// Package metrics exposes Prometheus metrics handler.
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
	categorySelectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_category_selected_total",
			Help: "Category selections grouped by source",
		},
		[]string{"source"},
	)
	llmSuggestionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_llm_suggestion_total",
			Help: "LLM suggestion result",
		},
		[]string{"result"},
	)
	mappingMutationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_mapping_mutation_total",
			Help: "Mapping mutation actions",
		},
		[]string{"action"},
	)
)

func init() {
	prometheus.MustRegister(updatesTotal)
	prometheus.MustRegister(transactionsSaved)
	prometheus.MustRegister(categorySelectedTotal)
	prometheus.MustRegister(llmSuggestionTotal)
	prometheus.MustRegister(mappingMutationTotal)
}

// IncUpdate increments updates counter.
func IncUpdate() { updatesTotal.Inc() }

// IncTransactionsSaved increments saved counter with a status label.
func IncTransactionsSaved(status string) { transactionsSaved.WithLabelValues(status).Inc() }
func IncCategorySelected(source string)  { categorySelectedTotal.WithLabelValues(source).Inc() }
func IncLLMSuggestion(result string)     { llmSuggestionTotal.WithLabelValues(result).Inc() }
func IncMappingMutation(action string)   { mappingMutationTotal.WithLabelValues(action).Inc() }

// Handler returns the HTTP handler for /metrics.
func Handler() http.Handler { return promhttp.Handler() }
