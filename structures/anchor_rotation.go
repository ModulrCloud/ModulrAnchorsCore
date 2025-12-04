package structures

import (
	"encoding/json"
	"fmt"
)

type AnchorRotationProof struct {
	EpochIndex int               `json:"epochIndex"`
	Anchor     string            `json:"anchor"`
	VotingStat VotingStat        `json:"votingStat"`
	Signatures map[string]string `json:"signatures"`
}

type LeaderFinalizationProof struct {
	EpochIndex int               `json:"epochIndex"`
	Leader     string            `json:"leader"`
	VotingStat VotingStat        `json:"votingStat"`
	Signatures map[string]string `json:"signatures"`
}

type BlockExtraData struct {
	Fields                   map[string]string         `json:"fields,omitempty"`
	RotationProofs           []AnchorRotationProof     `json:"rotationProofs,omitempty"`
	LeaderFinalizationProofs []LeaderFinalizationProof `json:"leaderFinalizationProofs,omitempty"`
}

type blockExtraDataAlias struct {
	Fields                   map[string]string         `json:"fields,omitempty"`
	RotationProofs           []AnchorRotationProof     `json:"rotationProofs,omitempty"`
	LeaderFinalizationProofs []LeaderFinalizationProof `json:"leaderFinalizationProofs,omitempty"`
}

func (extra BlockExtraData) MarshalJSON() ([]byte, error) {
	if len(extra.RotationProofs) == 0 && len(extra.LeaderFinalizationProofs) == 0 {
		if len(extra.Fields) == 0 {
			return []byte("{}"), nil
		}
		return json.Marshal(extra.Fields)
	}
	alias := blockExtraDataAlias(extra)
	if alias.Fields == nil {
		alias.Fields = map[string]string{}
	}
	return json.Marshal(alias)
}

func (extra *BlockExtraData) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*extra = BlockExtraData{}
		return nil
	}
	var alias blockExtraDataAlias
	if err := json.Unmarshal(data, &alias); err == nil && (alias.Fields != nil || alias.RotationProofs != nil || alias.LeaderFinalizationProofs != nil) {
		*extra = BlockExtraData(alias)
		return nil
	}
	var fields map[string]string
	if err := json.Unmarshal(data, &fields); err == nil {
		extra.Fields = fields
		extra.RotationProofs = nil
		extra.LeaderFinalizationProofs = nil
		return nil
	}
	return fmt.Errorf("invalid extraData payload")
}
