package cache

import (
	"sync/atomic"
)

var (
	hits   uint64
	misses uint64
)

// RecordHit records a cache hit
func RecordHit() {
	atomic.AddUint64(&hits, 1)
}

// RecordMiss records a cache miss
func RecordMiss() {
	atomic.AddUint64(&misses, 1)
}

// GetStats returns cache statistics
func GetStats() (uint64, uint64) {
	return atomic.LoadUint64(&hits), atomic.LoadUint64(&misses)
}

// GetHitRate returns the cache hit rate
func GetHitRate() float64 {
	totalHits := atomic.LoadUint64(&hits)
	totalMisses := atomic.LoadUint64(&misses)
	total := totalHits + totalMisses

	if total == 0 {
		return 0.0
	}

	return float64(totalHits) / float64(total)
}
