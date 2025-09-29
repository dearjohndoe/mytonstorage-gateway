package remotetonstorage

import (
	"strings"
	"sync"
	"time"

	tonstorage "github.com/xssnick/tonutils-storage/storage"
)

type BagsCacheConfig struct {
	MaxCacheEntries int
}

type torrentCacheEntry struct {
	torrent    *tonstorage.Torrent
	downloader tonstorage.TorrentDownloader
	bagSize    uint64
	lastUsed   time.Time
}

type BagsCache struct {
	cache   map[string]*torrentCacheEntry
	mutex   sync.RWMutex
	config  BagsCacheConfig
	metrics *RemoteTONStorageMetrics
}

func (bc *BagsCache) Get(bagID string) (*tonstorage.Torrent, tonstorage.TorrentDownloader, bool) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	entry, exists := bc.cache[strings.ToLower(bagID)]
	if !exists {
		if bc.metrics != nil {
			bc.metrics.cacheMisses.Inc()
		}
		return nil, nil, false
	}

	if bc.metrics != nil {
		bc.metrics.cacheHits.Inc()
	}

	entry.lastUsed = time.Now()
	return entry.torrent, entry.downloader, true
}

func (bc *BagsCache) Set(bagID string, torrent *tonstorage.Torrent, downloader tonstorage.TorrentDownloader) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	var bagSize uint64
	if torrent.Info != nil {
		bagSize = torrent.Info.FileSize
	}

	entry := &torrentCacheEntry{
		torrent:    torrent,
		downloader: downloader,
		bagSize:    bagSize,
		lastUsed:   time.Now(),
	}
	if bc.metrics != nil {
		bc.metrics.cacheHits.Inc()
	}

	bc.cache[strings.ToLower(bagID)] = entry

	for bc.freeUnsafe() {
	}

	if bc.metrics != nil {
		bc.metrics.activeTorrents.Set(float64(len(bc.cache)))
	}
}

func (bc *BagsCache) Clear() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	for _, entry := range bc.cache {
		if entry.downloader != nil {
			if bc.metrics != nil {
				bc.metrics.cacheHits.Inc()
			}
			entry.downloader.Close()
		}
	}

	bc.cache = make(map[string]*torrentCacheEntry, bc.config.MaxCacheEntries)
	if bc.metrics != nil {
		bc.metrics.activeTorrents.Set(0)
	}
}

func (bc *BagsCache) freeUnsafe() (updated bool) {
	if len(bc.cache) > bc.config.MaxCacheEntries {
		var oldestBagID string
		oldestTime := time.Now()

		for bagID, entry := range bc.cache {
			if entry.lastUsed.Before(oldestTime) {
				oldestBagID = bagID
				oldestTime = entry.lastUsed
			}
		}

		if oldestBagID != "" {
			if entry := bc.cache[oldestBagID]; entry != nil && entry.downloader != nil {
				entry.torrent.Stop()
				entry.downloader.Close()
			}
			delete(bc.cache, oldestBagID)
			if bc.metrics != nil {
				bc.metrics.cacheEvicts.Inc()
			}
			updated = true
		}
	}

	return
}

func NewBagsCache(maxCacheEntries int) *BagsCache {
	if maxCacheEntries <= 0 {
		maxCacheEntries = 100
	}

	return &BagsCache{
		cache: make(map[string]*torrentCacheEntry, maxCacheEntries),
		config: BagsCacheConfig{
			MaxCacheEntries: maxCacheEntries,
		},
	}
}

func (bc *BagsCache) WithMetrics(m *RemoteTONStorageMetrics) *BagsCache {
	bc.metrics = m
	return bc
}
