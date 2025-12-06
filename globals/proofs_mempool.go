package globals

import (
	"fmt"
	"sync"

	"github.com/modulrcloud/modulr-anchors-core/structures"
)

// Mempool to store two types of proofs:

var mempool = struct {
	sync.Mutex
	aggregatedAnchorRotationProofs     map[string]structures.AggregatedAnchorRotaionProof      // proof for modulr-anchors-core logic to rotate anchors on demand
	aggregatedLeaderFinalizationProofs map[string]structures.AggregatedLeaderFinalizationProof // proof for modulr-core logic to finalize last block by leader
}{
	aggregatedAnchorRotationProofs:     make(map[string]structures.AggregatedAnchorRotaionProof),
	aggregatedLeaderFinalizationProofs: make(map[string]structures.AggregatedLeaderFinalizationProof),
}

func anchorRotationProofMempoolKey(proof structures.AggregatedAnchorRotaionProof) string {
	return fmt.Sprintf("%d:%s:%d", proof.EpochIndex, proof.Anchor, proof.VotingStat.Index)
}

func leaderFinalizationProofMempoolKey(proof structures.AggregatedLeaderFinalizationProof) string {
	return fmt.Sprintf("%d:%s:%d", proof.EpochIndex, proof.Leader, proof.VotingStat.Index)
}

func AddAnchorRotationProofToMempool(proof structures.AggregatedAnchorRotaionProof) {

	mempool.Lock()

	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}

	mempool.aggregatedAnchorRotationProofs[anchorRotationProofMempoolKey(proof)] = proof
	mempool.Unlock()

}

func AddLeaderFinalizationProofToMempool(proof structures.AggregatedLeaderFinalizationProof) {

	mempool.Lock()

	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}

	mempool.aggregatedLeaderFinalizationProofs[leaderFinalizationProofMempoolKey(proof)] = proof
	mempool.Unlock()

}

func DrainAnchorRotationProofsFromMempool() []structures.AggregatedAnchorRotaionProof {

	mempool.Lock()
	defer mempool.Unlock()

	if len(mempool.aggregatedAnchorRotationProofs) == 0 {
		return nil
	}

	proofs := make([]structures.AggregatedAnchorRotaionProof, 0, len(mempool.aggregatedAnchorRotationProofs))

	for _, proof := range mempool.aggregatedAnchorRotationProofs {
		proofs = append(proofs, proof)
	}

	mempool.aggregatedAnchorRotationProofs = make(map[string]structures.AggregatedAnchorRotaionProof)

	return proofs

}

func DrainLeaderFinalizationProofsFromMempool() []structures.AggregatedLeaderFinalizationProof {

	mempool.Lock()
	defer mempool.Unlock()

	if len(mempool.aggregatedLeaderFinalizationProofs) == 0 {
		return nil
	}

	proofs := make([]structures.AggregatedLeaderFinalizationProof, 0, len(mempool.aggregatedLeaderFinalizationProofs))

	for _, proof := range mempool.aggregatedLeaderFinalizationProofs {
		proofs = append(proofs, proof)
	}

	mempool.aggregatedLeaderFinalizationProofs = make(map[string]structures.AggregatedLeaderFinalizationProof)

	return proofs

}
