package remotetonstorage

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RemoteTONStorageMetrics holds Prometheus collectors for remote TON storage client.
type RemoteTONStorageMetrics struct {
	cacheHits      prometheus.Counter
	cacheMisses    prometheus.Counter
	cacheEvicts    prometheus.Counter
	activeTorrents prometheus.Gauge

	downloaderCreations        *prometheus.CounterVec
	downloaderCreationDuration *prometheus.HistogramVec

	listFilesReqs     *prometheus.CounterVec
	listFilesDuration *prometheus.HistogramVec

	streamFileReqs     *prometheus.CounterVec
	streamFileDuration *prometheus.HistogramVec
	streamFileTTFB     *prometheus.HistogramVec
	streamFileBytes    prometheus.Counter
	activeStreams      prometheus.Gauge
}

func NewRemoteTONStorageMetrics(namespace, subsystem string) *RemoteTONStorageMetrics {
	m := &RemoteTONStorageMetrics{
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_hits_total",
			Help:      "Remote TON storage torrent cache hits.",
		}),
		cacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_misses_total",
			Help:      "Remote TON storage torrent cache misses.",
		}),
		cacheEvicts: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_evictions_total",
			Help:      "Remote TON storage torrent cache evictions.",
		}),
		activeTorrents: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "active_torrents",
			Help:      "Current number of active (cached) torrents.",
		}),

		downloaderCreations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "downloader_creations_total",
			Help:      "Downloader creations result count.",
		}, []string{"result"}),
		downloaderCreationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "downloader_creation_duration_seconds",
			Help:      "Downloader creation latency.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 8, 12, 20},
		}, []string{"result"}),

		listFilesReqs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "list_files_requests_total",
			Help:      "ListFiles requests count.",
		}, []string{"result"}),
		listFilesDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "list_files_duration_seconds",
			Help:      "ListFiles request duration.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 8, 12},
		}, []string{"result"}),

		streamFileReqs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "stream_file_requests_total",
			Help:      "StreamFile requests count.",
		}, []string{"result"}),
		streamFileDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "stream_file_duration_seconds",
			Help:      "Full stream duration from request start to Close().",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 8, 12, 20, 30, 45, 60},
		}, []string{"result"}),
		streamFileTTFB: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "stream_file_ttfb_seconds",
			Help:      "Time to first byte for file stream.",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.75, 1, 1.5, 2, 3, 5},
		}, []string{"result"}),
		streamFileBytes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "stream_file_bytes_total",
			Help:      "Total bytes streamed to clients.",
		}),
		activeStreams: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "active_streams",
			Help:      "Current number of active file streams.",
		}),
	}

	prometheus.MustRegister(
		m.cacheHits, m.cacheMisses, m.cacheEvicts, m.activeTorrents,
		m.downloaderCreations, m.downloaderCreationDuration,
		m.listFilesReqs, m.listFilesDuration,
		m.streamFileReqs, m.streamFileDuration, m.streamFileTTFB, m.streamFileBytes, m.activeStreams,
	)

	return m
}
