package memory

import "time"

type MemoryConfig struct {
	CleanupInterval time.Duration
}

const DefaultCleanupInterval = 10 * time.Minute
