package utils

import (
	"encoding/json"
	"strconv"

	"github.com/modulrcloud/modulr-anchors-core/databases"
)

// BlockCreatorHealthStatus stores metadata about why we stopped generating proofs for a creator.
type BlockCreatorHealthStatus struct {
	Epoch   int    `json:"epoch"`
	Creator string `json:"creator"`
}

func buildBlockCreatorHealthKey(epochID int, creator string) []byte {
	return []byte("BLOCK_CREATOR_HEALTH:" + strconv.Itoa(epochID) + ":" + creator)
}

// DisableFinalizationProofsForCreator stores a persistent flag to stop generating proofs for the creator.
func DisableFinalizationProofsForCreator(epochID int, creator string) error {
	status := BlockCreatorHealthStatus{
		Epoch:   epochID,
		Creator: creator,
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return databases.FINALIZATION_VOTING_STATS.Put(buildBlockCreatorHealthKey(epochID, creator), payload, nil)
}

// IsFinalizationProofsDisabled checks if the creator is banned for the provided epoch.
func IsFinalizationProofsDisabled(epochID int, creator string) bool {
	if _, err := databases.FINALIZATION_VOTING_STATS.Get(buildBlockCreatorHealthKey(epochID, creator), nil); err == nil {
		return true
	}
	return false
}
