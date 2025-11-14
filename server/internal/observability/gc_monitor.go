package observability

import (
	"context"
	"runtime"
	"time"

	"github.com/go-logr/logr"
)

// StartGCMonitor starts a goroutine that periodically samples GC statistics
// and records GC pause durations to the Prometheus histogram.
// The monitor runs until the context is cancelled or the stop channel is closed.
// Returns a channel that can be closed to stop the monitor.
func StartGCMonitor(ctx context.Context, interval time.Duration, logger logr.Logger) chan struct{} {
	stopChan := make(chan struct{})
	
	go func() {
		var lastPauseTotalNs uint64
		var lastNumGC uint32
		
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		// Initial sample to establish baseline
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		lastPauseTotalNs = memStats.PauseTotalNs
		lastNumGC = memStats.NumGC
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-stopChan:
				return
			case <-ticker.C:
				// Sample GC stats
				runtime.ReadMemStats(&memStats)
				
				// Calculate GC pause delta
				currentPauseTotalNs := memStats.PauseTotalNs
				currentNumGC := memStats.NumGC
				
				// Only record if GC occurred
				if currentNumGC > lastNumGC {
					// Calculate total pause time since last sample
					pauseDeltaNs := currentPauseTotalNs - lastPauseTotalNs
					gcCount := currentNumGC - lastNumGC
					
					// Record average pause per GC cycle (if multiple GCs occurred)
					if gcCount > 0 && pauseDeltaNs > 0 {
						avgPauseNs := pauseDeltaNs / uint64(gcCount)
						avgPauseSeconds := float64(avgPauseNs) / 1e9 // Convert nanoseconds to seconds
						
						// Record to histogram
						if histogram := GetGCPauseHistogram(); histogram != nil {
							histogram.Observe(avgPauseSeconds)
						}
					}
					
					// Update baseline
					lastPauseTotalNs = currentPauseTotalNs
					lastNumGC = currentNumGC
				}
			}
		}
	}()
	
	return stopChan
}

