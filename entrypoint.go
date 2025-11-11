package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ModulrCloud/ModulrAnchorsCore/databases"
	"github.com/ModulrCloud/ModulrAnchorsCore/globals"
	"github.com/ModulrCloud/ModulrAnchorsCore/handlers"
	"github.com/ModulrCloud/ModulrAnchorsCore/http_pack"
	"github.com/ModulrCloud/ModulrAnchorsCore/structures"
	"github.com/ModulrCloud/ModulrAnchorsCore/threads"
	"github.com/ModulrCloud/ModulrAnchorsCore/utils"
	"github.com/ModulrCloud/ModulrAnchorsCore/websocket_pack"

	"github.com/syndtr/goleveldb/leveldb"
)

func RunAnchorsChains() {

	if err := prepareAnchorsChains(); err != nil {

		utils.LogWithTime(fmt.Sprintf("Failed to prepare blockchain: %v", err), utils.RED_COLOR)

		utils.GracefulShutdown()

		return

	}

	//_________________________ RUN SEVERAL LOGICAL THREADS _________________________

	//✅ 1.Thread to find AEFPs and change the epoch for AT
	go threads.EpochRotationThread()

	//✅ 2.Share our blocks within quorum members and get the finalization proofs
	go threads.BlocksSharingAndProofsGrabingThread()

	//✅ 4.Start to generate blocks
	go threads.BlocksGenerationThread()

	//✅ 5.Start a separate thread to work with voting for blocks in a sync way (for security)
	go threads.LeaderRotationThread()

	//✅ 10.Start monitor anchors health
	go threads.AnchorsHealthChecker()

	//___________________ RUN SERVERS - WEBSOCKET AND HTTP __________________

	// Set the atomic flag to true

	globals.FLOOD_PREVENTION_FLAG_FOR_ROUTES.Store(true)

	go websocket_pack.CreateWebsocketServer()

	http_pack.CreateHTTPServer()

}

func prepareAnchorsChains() error {

	if info, err := os.Stat(globals.CHAINDATA_PATH); err != nil {

		if os.IsNotExist(err) {

			if err := os.MkdirAll(globals.CHAINDATA_PATH, 0755); err != nil {

				return fmt.Errorf("create chaindata directory: %w", err)

			}

		} else {

			return fmt.Errorf("check chaindata directory: %w", err)

		}

	} else if !info.IsDir() {

		return fmt.Errorf("chaindata path %s exists and is not a directory", globals.CHAINDATA_PATH)

	}

	databases.BLOCKS = utils.OpenDb("BLOCKS")
	databases.STATE = utils.OpenDb("STATE")
	databases.EPOCH_DATA = utils.OpenDb("EPOCH_DATA")
	databases.APPROVEMENT_THREAD_METADATA = utils.OpenDb("APPROVEMENT_THREAD_METADATA")
	databases.FINALIZATION_VOTING_STATS = utils.OpenDb("FINALIZATION_VOTING_STATS")

	// Load GT - Generation Thread handler
	if data, err := databases.BLOCKS.Get([]byte("GT"), nil); err == nil {

		var gtHandler structures.GenerationThreadMetadataHandler

		if err := json.Unmarshal(data, &gtHandler); err != nil {
			return fmt.Errorf("unmarshal GENERATION_THREAD metadata: %w", err)
		}

		handlers.GENERATION_THREAD_METADATA = gtHandler

	} else {

		handlers.GENERATION_THREAD_METADATA = structures.GenerationThreadMetadataHandler{
			EpochFullId: utils.Blake3("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+globals.GENESIS.NetworkId) + "#-1",
			PrevHash:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			NextIndex:   0,
		}

	}

	if data, err := databases.APPROVEMENT_THREAD_METADATA.Get([]byte("AT"), nil); err == nil {

		var atHandler structures.ApprovementThreadMetadataHandler

		if err := json.Unmarshal(data, &atHandler); err != nil {
			return fmt.Errorf("unmarshal APPROVEMENT_THREAD metadata: %w", err)
		}

		if atHandler.ValidatorsStoragesCache == nil {
			atHandler.ValidatorsStoragesCache = make(map[string]*structures.ValidatorStorage)
		}

		handlers.APPROVEMENT_THREAD_METADATA.Handler = atHandler

	}

	if handlers.APPROVEMENT_THREAD_METADATA.Handler.CoreMajorVersion == -1 {

		if err := setGenesisToState(); err != nil {
			return fmt.Errorf("write genesis to state: %w", err)
		}

		serializedApprovementThread, err := json.Marshal(handlers.APPROVEMENT_THREAD_METADATA.Handler)

		if err != nil {
			return fmt.Errorf("marshal APPROVEMENT_THREAD metadata: %w", err)
		}

		if err := databases.APPROVEMENT_THREAD_METADATA.Put([]byte("AT"), serializedApprovementThread, nil); err != nil {
			return fmt.Errorf("save APPROVEMENT_THREAD metadata: %w", err)
		}

	}

	return nil
}

func setGenesisToState() error {

	approvementThreadBatch := new(leveldb.Batch)

	epochTimestamp := globals.GENESIS.FirstEpochStartTimestamp

	validatorsRegistryForEpochHandler := []string{}

	validatorsRegistryForEpochHandler2 := []string{}

	// __________________________________ Load info about validators __________________________________

	for _, validatorStorage := range globals.GENESIS.Validators {

		validatorPubkey := validatorStorage.Pubkey

		serializedStorage, err := json.Marshal(validatorStorage)

		if err != nil {
			return err
		}

		approvementThreadBatch.Put([]byte(validatorPubkey+"_VALIDATOR_STORAGE"), serializedStorage)

		validatorsRegistryForEpochHandler = append(validatorsRegistryForEpochHandler, validatorPubkey)

		validatorsRegistryForEpochHandler2 = append(validatorsRegistryForEpochHandler2, validatorPubkey)

	}

	handlers.APPROVEMENT_THREAD_METADATA.Handler.NetworkParameters = globals.GENESIS.NetworkParameters.CopyNetworkParameters()

	// Commit changes
	if err := databases.APPROVEMENT_THREAD_METADATA.Write(approvementThreadBatch, nil); err != nil {
		return err
	}

	hashInput := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" + globals.GENESIS.NetworkId

	initEpochHash := utils.Blake3(hashInput)

	epochHandlerForApprovementThread := structures.EpochDataHandler{
		Id:                 0,
		Hash:               initEpochHash,
		ValidatorsRegistry: validatorsRegistryForEpochHandler,
		StartTimestamp:     epochTimestamp,
		Quorum:             []string{}, // will be assigned
		LeadersSequence:    []string{}, // will be assigned
		CurrentLeaderIndex: 0,
	}

	// Assign quorum - pseudorandomly and in deterministic way

	epochHandlerForApprovementThread.Quorum = utils.GetCurrentEpochQuorum(&epochHandlerForApprovementThread, handlers.APPROVEMENT_THREAD_METADATA.Handler.NetworkParameters.QuorumSize, initEpochHash)

	// Now set the block generators for epoch pseudorandomly and in deterministic way

	utils.SetLeadersSequence(&epochHandlerForApprovementThread, initEpochHash)

	handlers.APPROVEMENT_THREAD_METADATA.Handler.EpochDataHandler = epochHandlerForApprovementThread

	// Store epoch data for API

	currentEpochDataHandler := handlers.APPROVEMENT_THREAD_METADATA.Handler.EpochDataHandler

	jsonedCurrentEpochDataHandler, _ := json.Marshal(currentEpochDataHandler)

	databases.EPOCH_DATA.Put([]byte("EPOCH_HANDLER:"+strconv.Itoa(currentEpochDataHandler.Id)), jsonedCurrentEpochDataHandler, nil)

	return nil

}
