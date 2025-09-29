package remotetonstorage

import (
	"io"
	"sync"
	"time"
)

// meteredStream wraps the file stream to record bytes, ttfb, duration and active stream gauge.
type meteredStream struct {
	io.ReadCloser
	start         time.Time
	metrics       *RemoteTONStorageMetrics
	firstByteOnce *sync.Once
	ttfbObserved  bool
	closed        sync.Once
}

func (m *meteredStream) Read(p []byte) (int, error) {
	n, err := m.ReadCloser.Read(p)
	if n > 0 && m.metrics != nil {
		m.firstByteOnce.Do(func() {
			m.metrics.streamFileTTFB.WithLabelValues("success").Observe(time.Since(m.start).Seconds())
			m.ttfbObserved = true
		})
		m.metrics.streamFileBytes.Add(float64(n))
	}
	return n, err
}

func (m *meteredStream) Close() error {
	var errRet error
	m.closed.Do(func() {
		errRet = m.ReadCloser.Close()
		if m.metrics != nil {
			m.metrics.streamFileDuration.WithLabelValues("success").Observe(time.Since(m.start).Seconds())
			m.metrics.activeStreams.Dec()
		}
	})
	return errRet
}
