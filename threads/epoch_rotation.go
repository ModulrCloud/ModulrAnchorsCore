package threads

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/modulrcloud/modulr-anchors-core/block_pack"
	"github.com/modulrcloud/modulr-anchors-core/cryptography"
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

			epochFullID := epochHandlerRef.Hash + "#" + strconv.Itoa(epochHandlerRef.Id)

			if !utils.SignalAboutEpochRotationExists(epochHandlerRef.Id) {

				// If epoch is not fresh - send the signal to persistent db that we finish it - not to create AFPs, ALRPs anymore
				keyValue := []byte("EPOCH_FINISH:" + strconv.Itoa(epochHandlerRef.Id))

				databases.FINALIZATION_VOTING_STATS.Put(keyValue, []byte("TRUE"), nil)

			}

			if utils.SignalAboutEpochRotationExists(epochHandlerRef.Id) {

				majority := utils.GetQuorumMajority(epochHandlerRef)

				if AEFP_AND_FIRST_BLOCK_DATA.Aefp != nil && AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash != "" {

					// 1. Fetch first block

					firstBlock := block_pack.GetBlock(epochHandlerRef.Id, AEFP_AND_FIRST_BLOCK_DATA.FirstBlockCreator, 0, epochHandlerRef)

					// 2. Compare hashes

					if firstBlock != nil && firstBlock.GetHash() == AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash {

						// 3. Verify that quorum agreed batch of delayed transactions

						latestBatchIndex := readLatestBatchIndex()

						var delayedTransactionsToExecute []map[string]string

						jsonedDelayedTxs, _ := json.Marshal(firstBlock.ExtraData.DelayedTransactionsBatch.DelayedTransactions)

						dataThatShouldBeSigned := "SIG_DELAYED_OPERATIONS:" + strconv.Itoa(epochHandlerRef.Id) + ":" + string(jsonedDelayedTxs)

						okSignatures := 0

						unique := make(map[string]bool)

						quorumMap := make(map[string]bool)

						for _, pk := range epochHandlerRef.Quorum {
							quorumMap[strings.ToLower(pk)] = true
						}

						for signerPubKey, signa := range firstBlock.ExtraData.DelayedTransactionsBatch.Proofs {

							isOK := cryptography.VerifySignature(dataThatShouldBeSigned, signerPubKey, signa)

							loweredPubKey := strings.ToLower(signerPubKey)

							quorumMember := quorumMap[loweredPubKey]

							if isOK && quorumMember && !unique[loweredPubKey] {

								unique[loweredPubKey] = true

								okSignatures++

							}

						}

						// 5. Finally - check if this batch has bigger index than already executed
						// 6. Only in case it's indeed new batch - execute it

						handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

						// Before acquiring .Lock() for modification, disable route reads.
						// This prevents HTTP/WebSocket handlers from calling RLock() during update,
						// avoiding a flood scenario where excessive reads delay the writer.
						// Existing readers will finish normally; new ones are rejected via this flag.

						globals.FLOOD_PREVENTION_FLAG_FOR_ROUTES.Store(false)

						handlers.APPROVEMENT_THREAD_METADATA.RWMutex.Lock()

						if okSignatures >= majority && int64(epochHandlerRef.Id) > latestBatchIndex {

							latestBatchIndex = int64(epochHandlerRef.Id)

							delayedTransactionsToExecute = firstBlock.ExtraData.DelayedTransactionsBatch.DelayedTransactions

						}

						keyBytes := []byte("EPOCH_HANDLER:" + strconv.Itoa(epochHandlerRef.Id))

						valBytes, _ := json.Marshal(epochHandlerRef)

						databases.EPOCH_DATA.Put(keyBytes, valBytes, nil)

						var daoVotingContractCalls, allTheRestContractCalls []map[string]string

						atomicBatch := new(leveldb.Batch)

						for _, delayedTransaction := range delayedTransactionsToExecute {

							if delayedTxType, ok := delayedTransaction["type"]; ok {

								if delayedTxType == "votingAccept" {

									daoVotingContractCalls = append(daoVotingContractCalls, delayedTransaction)

								} else {

									allTheRestContractCalls = append(allTheRestContractCalls, delayedTransaction)

								}

							}

						}

						for key, value := range handlers.APPROVEMENT_THREAD_METADATA.Handler.ValidatorsStoragesCache {

							valBytes, _ := json.Marshal(value)

							atomicBatch.Put([]byte(key), valBytes)

						}

						utils.LogWithTime("Delayed txs were executed for epoch on AT: "+epochFullID, utils.GREEN_COLOR)

						//_______________________ Update the values for new epoch _______________________

						// Now, after the execution we can change the epoch id and get the new hash + prepare new temporary object

						nextEpochId := epochHandlerRef.Id + 1

						nextEpochHash := utils.Blake3(AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash)

						nextEpochQuorumSize := handlers.APPROVEMENT_THREAD_METADATA.Handler.NetworkParameters.QuorumSize

						nextEpochHandler := structures.EpochDataHandler{
							Id:              nextEpochId,
							Hash:            nextEpochHash,
							AnchorsRegistry: epochHandlerRef.AnchorsRegistry,
							Quorum:          utils.GetCurrentEpochQuorum(epochHandlerRef, nextEpochQuorumSize, nextEpochHash),
							StartTimestamp:  epochHandlerRef.StartTimestamp + uint64(handlers.APPROVEMENT_THREAD_METADATA.Handler.NetworkParameters.EpochDuration),
						}

						nextEpochDataHandler := structures.NextEpochDataHandler{
							NextEpochHash:               nextEpochHash,
							NextEpochValidatorsRegistry: nextEpochHandler.AnchorsRegistry,
							NextEpochQuorum:             nextEpochHandler.Quorum,
							NextEpochLeadersSequence:    nextEpochHandler.LeadersSequence,
						}

						jsonedNextEpochDataHandler, _ := json.Marshal(nextEpochDataHandler)

						atomicBatch.Put([]byte("EPOCH_DATA:"+strconv.Itoa(nextEpochId)), jsonedNextEpochDataHandler)

						// Finally - assign new handler

						handlers.APPROVEMENT_THREAD_METADATA.Handler.EpochDataHandler = nextEpochHandler

						// And commit all the changes on AT as a single atomic batch

						jsonedHandler, _ := json.Marshal(handlers.APPROVEMENT_THREAD_METADATA.Handler)

						atomicBatch.Put([]byte("AT"), jsonedHandler)

						// Clean cache

						clear(handlers.APPROVEMENT_THREAD_METADATA.Handler.ValidatorsStoragesCache)

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

			} else {

				handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

			}

		} else {

			handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

		}

		time.Sleep(200 * time.Millisecond)

	}

}
