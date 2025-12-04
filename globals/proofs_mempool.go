package globals

import (
	"fmt"
	"sync"

	"github.com/modulrcloud/modulr-anchors-core/structures"
)

var MEMPOOL = struct {
	sync.Mutex
	anchorsRotationProofs     map[string]structures.AnchorRotationProof
	leadersFinalizationProofs map[string]structures.LeaderFinalizationProof
}{
	anchorsRotationProofs:     make(map[string]structures.AnchorRotationProof),
	leadersFinalizationProofs: make(map[string]structures.LeaderFinalizationProof),
}

func anchorRotationProofMempoolKey(proof structures.AnchorRotationProof) string {
	return fmt.Sprintf("%d:%s:%d", proof.EpochIndex, proof.Anchor, proof.VotingStat.Index)
}

func leaderFinalizationProofMempoolKey(proof structures.LeaderFinalizationProof) string {
	return fmt.Sprintf("%d:%s:%d", proof.EpochIndex, proof.Leader, proof.VotingStat.Index)
}

func AddAnchorRotationProofToMempool(proof structures.AnchorRotationProof) {

	MEMPOOL.Lock()

	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}

	MEMPOOL.anchorsRotationProofs[anchorRotationProofMempoolKey(proof)] = proof
	MEMPOOL.Unlock()

}

func AddLeaderFinalizationProofToMempool(proof structures.LeaderFinalizationProof) {

	MEMPOOL.Lock()

	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}

	MEMPOOL.leadersFinalizationProofs[leaderFinalizationProofMempoolKey(proof)] = proof
	MEMPOOL.Unlock()

}

func DrainRotationProofsFromMempool() []structures.AnchorRotationProof {

	MEMPOOL.Lock()
	defer MEMPOOL.Unlock()

	if len(MEMPOOL.anchorsRotationProofs) == 0 {
		return nil
	}

	proofs := make([]structures.AnchorRotationProof, 0, len(MEMPOOL.anchorsRotationProofs))

	for _, proof := range MEMPOOL.anchorsRotationProofs {
		proofs = append(proofs, proof)
	}

	MEMPOOL.anchorsRotationProofs = make(map[string]structures.AnchorRotationProof)

	return proofs

}

func DrainLeaderFinalizationProofsFromMempool() []structures.LeaderFinalizationProof {

	MEMPOOL.Lock()
	defer MEMPOOL.Unlock()

	if len(MEMPOOL.leadersFinalizationProofs) == 0 {
		return nil
	}

	proofs := make([]structures.LeaderFinalizationProof, 0, len(MEMPOOL.leadersFinalizationProofs))

	for _, proof := range MEMPOOL.leadersFinalizationProofs {
		proofs = append(proofs, proof)
	}

	MEMPOOL.leadersFinalizationProofs = make(map[string]structures.LeaderFinalizationProof)

	return proofs

}
