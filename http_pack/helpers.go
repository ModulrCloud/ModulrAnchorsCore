package http_pack

import (
	"github.com/modulrcloud/modulr-anchors-core/handlers"
	"github.com/modulrcloud/modulr-anchors-core/structures"
)

func getEpochHandlerByID(id int) *structures.EpochDataHandler {
	handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RLock()
	defer handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()
	epochHandlers := handlers.APPROVEMENT_THREAD_METADATA.Handler.GetEpochHandlers()
	for idx := range epochHandlers {
		if epochHandlers[idx].Id == id {
			return &epochHandlers[idx]
		}
	}
	return nil
}

func creatorInEpoch(creator string, registry []string) bool {
	for _, candidate := range registry {
		if candidate == creator {
			return true
		}
	}
	return false
}
