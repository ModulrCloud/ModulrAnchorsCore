package threads

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/globals"
	"github.com/modulrcloud/modulr-anchors-core/handlers"
	"github.com/modulrcloud/modulr-anchors-core/structures"
	"github.com/modulrcloud/modulr-anchors-core/utils"

	"github.com/syndtr/goleveldb/leveldb"
)

func EpochRotationThread() {

	for {

		handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RLock()

		if !utils.EpochStillFresh(&handlers.APPROVEMENT_THREAD_METADATA.Handler) {

			epochHandlerRef := &handlers.APPROVEMENT_THREAD_METADATA.Handler.EpochDataHandler

			if !utils.SignalAboutEpochRotationExists(epochHandlerRef.Id) {

				// If epoch is not fresh - send the signal to persistent db that we finish it - not to create AFPs, ALRPs anymore
				keyValue := []byte("EPOCH_FINISH:" + strconv.Itoa(epochHandlerRef.Id))

				databases.FINALIZATION_VOTING_STATS.Put(keyValue, []byte("TRUE"), nil)

			}

			if utils.SignalAboutEpochRotationExists(epochHandlerRef.Id) {

				handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

				// Before acquiring .Lock() for modification, disable route reads.
				// This prevents HTTP/WebSocket handlers from calling RLock() during update,
				// avoiding a flood scenario where excessive reads delay the writer.
				// Existing readers will finish normally; new ones are rejected via this flag.

				globals.FLOOD_PREVENTION_FLAG_FOR_ROUTES.Store(false)

				handlers.APPROVEMENT_THREAD_METADATA.RWMutex.Lock()

				keyBytes := []byte("EPOCH_HANDLER:" + strconv.Itoa(epochHandlerRef.Id))

				valBytes, _ := json.Marshal(epochHandlerRef)

				databases.EPOCH_DATA.Put(keyBytes, valBytes, nil)

				atomicBatch := new(leveldb.Batch)

				//_______________________ Update the values for new epoch _______________________

				// Now, after the execution we can change the epoch id and get the new hash + prepare new temporary object

				nextEpochId := epochHandlerRef.Id + 1

				nextEpochHash := utils.Blake3(epochHandlerRef.Hash)

				nextEpochQuorumSize := handlers.APPROVEMENT_THREAD_METADATA.Handler.NetworkParameters.QuorumSize

				nextEpochHandler := structures.EpochDataHandler{
					Id:              nextEpochId,
					Hash:            nextEpochHash,
					AnchorsRegistry: epochHandlerRef.AnchorsRegistry,
					Quorum:          utils.GetCurrentEpochQuorum(epochHandlerRef, nextEpochQuorumSize, nextEpochHash),
					StartTimestamp:  epochHandlerRef.StartTimestamp + uint64(handlers.APPROVEMENT_THREAD_METADATA.Handler.NetworkParameters.EpochDuration),
				}

				// Finally - assign new handler

				handlers.APPROVEMENT_THREAD_METADATA.Handler.EpochDataHandler = nextEpochHandler

				// And commit all the changes on AT as a single atomic batch

				jsonedHandler, _ := json.Marshal(handlers.APPROVEMENT_THREAD_METADATA.Handler)

				atomicBatch.Put([]byte("AT"), jsonedHandler)

				if batchCommitErr := databases.APPROVEMENT_THREAD_METADATA.Write(atomicBatch, nil); batchCommitErr != nil {

					panic("Error with writing batch to approvement thread db. Try to launch again")

				}

				utils.LogWithTime("Epoch on approvement thread was updated => "+nextEpochHash+"#"+strconv.Itoa(nextEpochId), utils.GREEN_COLOR)

				handlers.APPROVEMENT_THREAD_METADATA.RWMutex.Unlock()

				// Re-enable route reads after modification is complete.
				// New HTTP/WebSocket handlers can now call RLock() as usual

				globals.FLOOD_PREVENTION_FLAG_FOR_ROUTES.Store(true)

			} else {

				handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

			}

		} else {

			handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

		}

		time.Sleep(200 * time.Millisecond)

	}

}
