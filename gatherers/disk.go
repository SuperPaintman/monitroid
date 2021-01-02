package gatherers

import (
	"fmt"

	"github.com/c9s/goprocinfo/linux"
)

type DiskStats struct {
	Usage float64 `json:"usage"`
}

var _ Gatherer = (*Disk)(nil)

type Disk struct{}

func (c *Disk) Gather() (interface{}, error) {
	stats, err := linux.ReadDisk("/")
	if err != nil {
		return nil, fmt.Errorf("gatherers: failed to read proc diskstats: %w", err)
	}

	usage := float64(stats.Used) / float64(stats.All)

	return &DiskStats{
		Usage: usage,
	}, nil
}
