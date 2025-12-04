package utils

import (
	"encoding/json"
	"errors"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/structures"

	ldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
)

func leaderFinalizationKey(chainId, leader string) []byte {
	return []byte("LEADER_FINALIZATION_PROOF:" + chainId + ":" + leader)
}

func leaderVotingStatKey(chainId, leader string) []byte {
	return []byte("LEADER_VOTING_STAT:" + chainId + ":" + leader)
}

func StoreLeaderFinalizationProof(proof structures.LeaderFinalizationProof) error {
	payload, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	return databases.FINALIZATION_VOTING_STATS.Put(leaderFinalizationKey(proof.ChainId, proof.Leader), payload, nil)
}

func LoadLeaderFinalizationProof(chainId, leader string) (structures.LeaderFinalizationProof, error) {
	var proof structures.LeaderFinalizationProof
	raw, err := databases.FINALIZATION_VOTING_STATS.Get(leaderFinalizationKey(chainId, leader), nil)
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

func StoreLeaderVotingStat(chainId, leader string, stat structures.VotingStat) error {
	payload, err := json.Marshal(stat)
	if err != nil {
		return err
	}
	return databases.FINALIZATION_VOTING_STATS.Put(leaderVotingStatKey(chainId, leader), payload, nil)
}

func LoadLeaderVotingStat(chainId, leader string) (structures.VotingStat, error) {
	stat := structures.NewVotingStatTemplate()
	raw, err := databases.FINALIZATION_VOTING_STATS.Get(leaderVotingStatKey(chainId, leader), nil)
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
