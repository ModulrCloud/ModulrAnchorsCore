package utils

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/structures"

	ldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
)

const rotationProofPrefix = "ANCHOR_ROTATION_PROOF:"

func rotationProofKey(epoch int, creator string) []byte {
	return []byte(rotationProofPrefix + strconv.Itoa(epoch) + ":" + creator)
}

func StoreRotationProof(proof structures.AnchorRotationProofBundle) error {
	payload, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	return databases.FINALIZATION_VOTING_STATS.Put(rotationProofKey(proof.EpochIndex, proof.Creator), payload, nil)
}

func LoadRotationProof(epoch int, creator string) (structures.AnchorRotationProofBundle, error) {
	var proof structures.AnchorRotationProofBundle
	raw, err := databases.FINALIZATION_VOTING_STATS.Get(rotationProofKey(epoch, creator), nil)
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

func HasRotationProof(epoch int, creator string) bool {
	if _, err := databases.FINALIZATION_VOTING_STATS.Get(rotationProofKey(epoch, creator), nil); err == nil {
		return true
	}
	return false
}
