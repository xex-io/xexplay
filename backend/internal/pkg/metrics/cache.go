package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// CacheHitsTotal counts total cache hits by key type.
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xexplay",
			Name:      "cache_hits_total",
			Help:      "Total number of cache hits.",
		},
		[]string{"key_type"},
	)

	// CacheMissesTotal counts total cache misses by key type.
	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xexplay",
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses.",
		},
		[]string{"key_type"},
	)
)
