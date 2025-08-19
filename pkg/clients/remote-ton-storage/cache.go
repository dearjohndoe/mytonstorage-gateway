package remotetonstorage

import (
	"strings"
	"sync"
	"time"

	tonstorage "github.com/xssnick/tonutils-storage/storage"
)

type BagsCacheConfig struct {
	MaxCacheSize    uint64
	MaxCacheEntries int
}

type torrentCacheEntry struct {
	torrent    *tonstorage.Torrent
	downloader tonstorage.TorrentDownloader
	bagSize    uint64
	lastUsed   time.Time
}

type BagsCache struct {
	cache  map[string]*torrentCacheEntry
	mutex  sync.RWMutex
	config BagsCacheConfig
}

func (bc *BagsCache) Get(bagID string) (*tonstorage.Torrent, tonstorage.TorrentDownloader, bool) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	entry, exists := bc.cache[strings.ToLower(bagID)]
	if !exists {
		return nil, nil, false
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

	bc.cache[strings.ToLower(bagID)] = entry

	for bc.freeUnsafe() {
	}
}

func (bc *BagsCache) Clear() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	for _, entry := range bc.cache {
		if entry.downloader != nil {
			entry.downloader.Close()
		}
	}

	bc.cache = make(map[string]*torrentCacheEntry, bc.config.MaxCacheEntries)
}

func (bc *BagsCache) freeUnsafe() (updated bool) {
	var totalSize uint64
	for _, entry := range bc.cache {
		totalSize += entry.bagSize
	}

	if totalSize > bc.config.MaxCacheSize || len(bc.cache) > bc.config.MaxCacheEntries {
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
				entry.downloader.Close()
			}
			delete(bc.cache, oldestBagID)
			updated = true
		}
	}

	return
}

func NewBagsCache(maxCacheSize uint64, maxCacheEntries int) *BagsCache {
	return &BagsCache{
		cache: make(map[string]*torrentCacheEntry, maxCacheEntries),
		config: BagsCacheConfig{
			MaxCacheSize:    maxCacheSize,
			MaxCacheEntries: maxCacheEntries,
		},
	}
}
