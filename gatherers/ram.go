package gatherers

import (
	"fmt"

	"github.com/c9s/goprocinfo/linux"
)

type RAMStats struct {
	Usage float64 `json:"usage"`
}

var _ Gatherer = (*RAM)(nil)

type RAM struct{}

func (c *RAM) Gather() (interface{}, error) {
	stats, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("gatherers: failed to read proc meminfo: %w", err)
	}

	usage := float64(stats.MemTotal-(stats.MemFree+stats.Buffers+stats.Cached+stats.InactiveAnon)) / float64(stats.MemTotal)

	return &RAMStats{
		Usage: usage,
	}, nil
}
