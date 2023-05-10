package common

import (
	"log"
	"time"
)

func DurationCheck(d time.Time) {
	duration := time.Since(d)
	minutes := int(duration.Minutes())
	seconds := duration.Seconds() - float64(minutes*60)
	log.Printf("[TIME]: %dm %.2fs\n", minutes, seconds)
}
