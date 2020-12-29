package gatherers

import (
	"fmt"
	"sync"

	"github.com/c9s/goprocinfo/linux"
)

type CPUStats struct {
	Usage float64 `json:"usage"`
}

var _ Gatherer = (*CPU)(nil)

type CPU struct {
	mu   sync.Mutex
	prev linux.CPUStat
}

func (c *CPU) Gather() (interface{}, error) {
	stats, err := linux.ReadStat("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("gatherers: failed to read proc stat: %w", err)
	}

	defer func() {
		c.mu.Lock()
		c.prev = stats.CPUStatAll
		c.mu.Unlock()
	}()

	used, total := calcUsedAndTotal(stats.CPUStatAll)

	c.mu.Lock()
	prevUsed, prevTotal := calcUsedAndTotal(c.prev)
	c.mu.Unlock()

	usedDiff := used - prevUsed
	totalDiff := total - prevTotal

	usage := float64(totalDiff-usedDiff) / float64(totalDiff)

	return &CPUStats{
		Usage: usage,
	}, nil
}

func calcUsedAndTotal(stat linux.CPUStat) (used uint64, total uint64) {
	used = stat.Idle + stat.IOWait
	total = stat.User + stat.Nice + stat.System + stat.Idle + stat.IOWait + stat.IRQ + stat.SoftIRQ + stat.Steal

	return
}
