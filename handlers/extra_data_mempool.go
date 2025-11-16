package handlers

import (
	"fmt"
	"sync"

	"github.com/modulrcloud/modulr-anchors-core/structures"
)

var extraDataMempool = struct {
	sync.Mutex
	rotationProofs map[string]structures.AnchorRotationProofBundle
}{rotationProofs: make(map[string]structures.AnchorRotationProofBundle)}

func rotationProofMempoolKey(proof structures.AnchorRotationProofBundle) string {
	return fmt.Sprintf("%d:%s:%d", proof.EpochIndex, proof.Creator, proof.VotingStat.Index)
}

func AddRotationProofToMempool(proof structures.AnchorRotationProofBundle) {
	extraDataMempool.Lock()
	if proof.Signatures == nil {
		proof.Signatures = map[string]string{}
	}
	extraDataMempool.rotationProofs[rotationProofMempoolKey(proof)] = proof
	extraDataMempool.Unlock()
}

func DrainRotationProofsFromMempool() []structures.AnchorRotationProofBundle {
	extraDataMempool.Lock()
	defer extraDataMempool.Unlock()
	if len(extraDataMempool.rotationProofs) == 0 {
		return nil
	}
	proofs := make([]structures.AnchorRotationProofBundle, 0, len(extraDataMempool.rotationProofs))
	for _, proof := range extraDataMempool.rotationProofs {
		proofs = append(proofs, proof)
	}
	extraDataMempool.rotationProofs = make(map[string]structures.AnchorRotationProofBundle)
	return proofs
}
