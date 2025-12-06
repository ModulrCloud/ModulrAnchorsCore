package utils

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/structures"

	ldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
)

func leaderFinalizationKey(epochIndex int, leader string) []byte {

	return []byte("LEADER_FINALIZATION_PROOF:" + strconv.Itoa(epochIndex) + ":" + leader)

}

func StoreLeaderFinalizationProof(proof structures.AggregatedLeaderFinalizationProof) error {

	payload, err := json.Marshal(proof)

	if err != nil {
		return err
	}

	return databases.FINALIZATION_VOTING_STATS.Put(leaderFinalizationKey(proof.EpochIndex, proof.Leader), payload, nil)

}

func LoadLeaderFinalizationProof(epochIndex int, leader string) (structures.AggregatedLeaderFinalizationProof, error) {

	var proof structures.AggregatedLeaderFinalizationProof

	raw, err := databases.FINALIZATION_VOTING_STATS.Get(leaderFinalizationKey(epochIndex, leader), nil)

	if err != nil {
		if errors.Is(err, ldbErrors.ErrNotFound) {
			return proof, nil
		}
		return proof, err
	}

	if len(raw) == 0 {
		return proof, nil
	}

	if err := json.Unmarshal(raw, &proof); err != nil {
		return proof, err
	}
	return proof, nil

}
