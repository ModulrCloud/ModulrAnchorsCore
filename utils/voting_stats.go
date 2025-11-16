package utils

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/structures"

	ldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
)

func BuildVotingStatKey(epochIndex int, creator string) []byte {
	return []byte(strconv.Itoa(epochIndex) + ":" + creator)
}

func ReadVotingStat(epochIndex int, creator string) (structures.VotingStat, error) {
	key := BuildVotingStatKey(epochIndex, creator)
	stat := structures.NewVotingStatTemplate()
	raw, err := databases.FINALIZATION_VOTING_STATS.Get(key, nil)
	if err != nil {
		if errors.Is(err, ldbErrors.ErrNotFound) {
			return stat, nil
		}
		return stat, err
	}
	if len(raw) == 0 {
		return stat, nil
	}
	if err := json.Unmarshal(raw, &stat); err != nil {
		return stat, err
	}
	return stat, nil
}

func StoreVotingStat(epochIndex int, creator string, stat structures.VotingStat) error {
	payload, err := json.Marshal(stat)
	if err != nil {
		return err
	}
	return databases.FINALIZATION_VOTING_STATS.Put(BuildVotingStatKey(epochIndex, creator), payload, nil)
}
