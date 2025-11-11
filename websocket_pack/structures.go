package websocket_pack

import (
	"github.com/ModulrCloud/ModulrAnchorsCore/block_pack"
	"github.com/ModulrCloud/ModulrAnchorsCore/structures"
)

type WsFinalizationProofRequest struct {
	Route            string                                 `json:"route"`
	Block            block_pack.Block                       `json:"block"`
	PreviousBlockAfp structures.AggregatedFinalizationProof `json:"previousBlockAfp"`
}

type WsFinalizationProofResponse struct {
	Voter             string `json:"voter"`
	FinalizationProof string `json:"finalizationProof"`
	VotedForHash      string `json:"votedForHash"`
}

type WsBlockWithAfpRequest struct {
	Route   string `json:"route"`
	BlockId string `json:"blockID"`
}

type WsBlockWithAfpResponse struct {
	Block *block_pack.Block                       `json:"block"`
	Afp   *structures.AggregatedFinalizationProof `json:"afp"`
}
