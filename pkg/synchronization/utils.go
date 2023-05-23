package synchronization

import (
	"strings"

	"github.com/mutagen-io/mutagen/pkg/synchronization/core"
)

func getSyncMode(syncMode string) core.SynchronizationMode {
	switch strings.ToLower(syncMode) {
	case "two-way-safe":
		return core.SynchronizationMode_SynchronizationModeTwoWaySafe
	case "two-way-resolved":
		return core.SynchronizationMode_SynchronizationModeTwoWayResolved
	case "one-way-safe":
		return core.SynchronizationMode_SynchronizationModeOneWaySafe
	case "one-way-replica":
		return core.SynchronizationMode_SynchronizationModeOneWayReplica
	}
	return core.SynchronizationMode_SynchronizationModeDefault
}
