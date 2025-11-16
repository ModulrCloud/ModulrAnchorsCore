package threads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/modulrcloud/modulr-anchors-core/globals"
	"github.com/modulrcloud/modulr-anchors-core/handlers"
	"github.com/modulrcloud/modulr-anchors-core/structures"
	"github.com/modulrcloud/modulr-anchors-core/utils"
)

const rotationCollectorInterval = 5 * time.Second

var httpClient = &http.Client{Timeout: 5 * time.Second}

func AnchorRotationCollectorThread() {
	ticker := time.NewTicker(rotationCollectorInterval)
	defer ticker.Stop()
	for range ticker.C {
		collectRotationProofs()
	}
}

func collectRotationProofs() {
	handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RLock()
	epochHandlers := handlers.APPROVEMENT_THREAD_METADATA.Handler.GetEpochHandlers()
	handlers.APPROVEMENT_THREAD_METADATA.RWMutex.RUnlock()

	for idx := range epochHandlers {
		handleEpochForRotation(&epochHandlers[idx])
	}
}

func handleEpochForRotation(epochHandler *structures.EpochDataHandler) {
	if len(epochHandler.AnchorsRegistry) == 0 {
		return
	}
	for _, creator := range epochHandler.AnchorsRegistry {
		processCreatorRotation(epochHandler, creator)
	}
}

func processCreatorRotation(epochHandler *structures.EpochDataHandler, creator string) {
	if !utils.IsFinalizationProofsDisabled(epochHandler.Id, creator) {
		return
	}
	if utils.HasRotationProof(epochHandler.Id, creator) {
		return
	}

	mutex := utils.GetBlockCreatorMutex(epochHandler.Id, creator)
	mutex.Lock()
	defer mutex.Unlock()

	if !utils.IsFinalizationProofsDisabled(epochHandler.Id, creator) || utils.HasRotationProof(epochHandler.Id, creator) {
		return
	}

	stat, err := utils.ReadVotingStat(epochHandler.Id, creator)
	if err != nil {
		utils.LogWithTime(fmt.Sprintf("anchor rotation: failed to read voting stat for %s in epoch %d: %v", creator, epochHandler.Id, err), utils.YELLOW_COLOR)
		return
	}
	if stat.Index < 0 || stat.Hash == "" {
		return
	}

	signatures := collectRotationSignatures(epochHandler, creator, stat)
	majority := utils.GetQuorumMajority(epochHandler)
	if len(signatures) < majority {
		return
	}

	proof := structures.AnchorRotationProofBundle{
		EpochIndex: epochHandler.Id,
		Creator:    creator,
		VotingStat: stat,
		Signatures: signatures,
	}
	if err := utils.StoreRotationProof(proof); err != nil {
		utils.LogWithTime(fmt.Sprintf("anchor rotation: failed to persist proof for %s epoch %d: %v", creator, epochHandler.Id, err), utils.YELLOW_COLOR)
		return
	}
	handlers.AddRotationProofToMempool(proof)
	broadcastRotationProof(epochHandler, proof)
	utils.LogWithTime(fmt.Sprintf("anchor rotation: collected %d signatures for %s in epoch %d", len(signatures), creator, epochHandler.Id), utils.GREEN_COLOR)
}

func collectRotationSignatures(epochHandler *structures.EpochDataHandler, creator string, stat structures.VotingStat) map[string]string {
	quorumMembers := utils.GetQuorumUrlsAndPubkeys(epochHandler)
	payload := structures.AnchorRotationProofRequest{EpochIndex: epochHandler.Id, Creator: creator, Proposal: stat}
	requestBody, _ := json.Marshal(payload)
	signatures := make(map[string]string)
	majority := utils.GetQuorumMajority(epochHandler)

	for _, member := range quorumMembers {
		if member.PubKey == globals.CONFIGURATION.PublicKey || member.Url == "" {
			continue
		}
		endpoint := strings.TrimRight(member.Url, "/") + "/anchor_rotation_proof"
		body, status, err := postJSON(endpoint, requestBody)
		if err != nil {
			continue
		}
		var response structures.AnchorRotationProofResponse
		if err := json.Unmarshal(body, &response); err != nil {
			continue
		}
		switch response.Status {
		case "UPGRADE":
			if response.VotingStat != nil {
				if err := utils.StoreVotingStat(epochHandler.Id, creator, *response.VotingStat); err != nil {
					utils.LogWithTime(fmt.Sprintf("anchor rotation: failed to store upgraded stat for %s epoch %d: %v", creator, epochHandler.Id, err), utils.YELLOW_COLOR)
				}
				return nil
			}
		case "OK":
			if response.VotingStat == nil {
				continue
			}
			if response.VotingStat.Index > stat.Index || !strings.EqualFold(response.VotingStat.Hash, stat.Hash) {
				if err := utils.StoreVotingStat(epochHandler.Id, creator, *response.VotingStat); err != nil {
					utils.LogWithTime(fmt.Sprintf("anchor rotation: failed to store fresher stat for %s epoch %d: %v", creator, epochHandler.Id, err), utils.YELLOW_COLOR)
				}
				return nil
			}
			if response.Signature != "" && status == http.StatusOK {
				signatures[member.PubKey] = response.Signature
			}
		}
		if len(signatures) >= majority {
			break
		}
	}
	return signatures
}

func postJSON(url string, payload []byte) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func broadcastRotationProof(epochHandler *structures.EpochDataHandler, proof structures.AnchorRotationProofBundle) {
	payload := structures.AcceptExtraDataRequest{RotationProofs: []structures.AnchorRotationProofBundle{proof}}
	body, _ := json.Marshal(payload)
	for _, member := range utils.GetQuorumUrlsAndPubkeys(epochHandler) {
		if member.PubKey == globals.CONFIGURATION.PublicKey || member.Url == "" {
			continue
		}
		endpoint := strings.TrimRight(member.Url, "/") + "/accept_extra_data"
		if _, _, err := postJSON(endpoint, body); err != nil {
			utils.LogWithTime(fmt.Sprintf("anchor rotation: failed to broadcast proof to %s: %v", member.PubKey, err), utils.YELLOW_COLOR)
		}
	}
}
