package globals

import (
	"fmt"
	"sync"

	"github.com/modulrcloud/modulr-anchors-core/structures"
)

var MEMPOOL = struct {
	sync.Mutex
	anchorsRotationProofs     map[string]structures.AnchorRotationProofBundle
	leadersFinalizationProofs map[string]structures.LeaderFinalizationProofBundle
}{
	anchorsRotationProofs:     make(map[string]structures.AnchorRotationProofBundle),
	leadersFinalizationProofs: make(map[string]structures.LeaderFinalizationProofBundle),
}

func anchorRotationProofMempoolKey(proof structures.AnchorRotationProofBundle) string {
	return fmt.Sprintf("%d:%s:%d", proof.EpochIndex, proof.Creator, proof.VotingStat.Index)
}

func leaderFinalizationProofMempoolKey(proof structures.LeaderFinalizationProofBundle) string {
	return fmt.Sprintf("%s:%s:%d", proof.ChainId, proof.Leader, proof.VotingStat.Index)
}

func AddAnchorRotationProofToMempool(proof structures.AnchorRotationProofBundle) {

	MEMPOOL.Lock()

	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}

	MEMPOOL.anchorsRotationProofs[anchorRotationProofMempoolKey(proof)] = proof
	MEMPOOL.Unlock()

}

func AddLeaderFinalizationProofToMempool(proof structures.LeaderFinalizationProofBundle) {

	MEMPOOL.Lock()

	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}

	MEMPOOL.leadersFinalizationProofs[leaderFinalizationProofMempoolKey(proof)] = proof
	MEMPOOL.Unlock()

}

func DrainRotationProofsFromMempool() []structures.AnchorRotationProofBundle {

	MEMPOOL.Lock()
	defer MEMPOOL.Unlock()

	if len(MEMPOOL.anchorsRotationProofs) == 0 {
		return nil
	}

	proofs := make([]structures.AnchorRotationProofBundle, 0, len(MEMPOOL.anchorsRotationProofs))

	for _, proof := range MEMPOOL.anchorsRotationProofs {
		proofs = append(proofs, proof)
	}

	MEMPOOL.anchorsRotationProofs = make(map[string]structures.AnchorRotationProofBundle)

	return proofs

}

func DrainLeaderFinalizationProofsFromMempool() []structures.LeaderFinalizationProofBundle {

	MEMPOOL.Lock()
	defer MEMPOOL.Unlock()

	if len(MEMPOOL.leadersFinalizationProofs) == 0 {
		return nil
	}

	proofs := make([]structures.LeaderFinalizationProofBundle, 0, len(MEMPOOL.leadersFinalizationProofs))

	for _, proof := range MEMPOOL.leadersFinalizationProofs {
		proofs = append(proofs, proof)
	}

	MEMPOOL.leadersFinalizationProofs = make(map[string]structures.LeaderFinalizationProofBundle)

	return proofs

}
