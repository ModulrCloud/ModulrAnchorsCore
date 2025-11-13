package block_pack

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modulrcloud/modulr-anchors-core/cryptography"
	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/globals"
	"github.com/modulrcloud/modulr-anchors-core/structures"
	"github.com/modulrcloud/modulr-anchors-core/utils"
)

type Block struct {
	Creator   string            `json:"creator"`
	Time      int64             `json:"time"`
	Epoch     string            `json:"epoch"`
	ExtraData map[string]string `json:"extraData"`
	Index     int               `json:"index"`
	PrevHash  string            `json:"prevHash"`
	Sig       string            `json:"sig"`
}

func formatExtraData(extraData map[string]string) string {
	if len(extraData) == 0 {
		return ""
	}

	keys := make([]string, 0, len(extraData))
	for key := range extraData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+extraData[key])
	}

	return strings.Join(parts, ",")
}

func NewBlock(extraData map[string]string, epochFullID string, metadata *structures.GenerationThreadMetadataHandler) *Block {
	return &Block{
		Creator:   globals.CONFIGURATION.PublicKey,
		Time:      utils.GetUTCTimestampInMilliSeconds(),
		Epoch:     epochFullID,
		ExtraData: extraData,
		Index:     metadata.NextIndex,
		PrevHash:  metadata.PrevHash,
		Sig:       "",
	}
}

func (block *Block) GetHash() string {

	dataToHash := strings.Join([]string{
		block.Creator,
		strconv.FormatInt(block.Time, 10),
		globals.GENESIS.NetworkId,
		block.Epoch,
		formatExtraData(block.ExtraData),
		strconv.Itoa(block.Index),
		block.PrevHash,
	}, ":")

	return utils.Blake3(dataToHash)
}

func (block *Block) SignBlock() {

	block.Sig = cryptography.GenerateSignature(globals.CONFIGURATION.PrivateKey, block.GetHash())

}

func (block *Block) VerifySignature() bool {

	return cryptography.VerifySignature(block.GetHash(), block.Creator, block.Sig)

}

func GetBlock(epochIndex int, blockCreator string, index uint, epochHandler *structures.EpochDataHandler) *Block {

	blockID := strconv.Itoa(epochIndex) + ":" + blockCreator + ":" + strconv.Itoa(int(index))

	blockAsBytes, err := databases.BLOCKS.Get([]byte(blockID), nil)

	if err == nil {

		var blockParsed *Block

		err = json.Unmarshal(blockAsBytes, &blockParsed)

		if err == nil {
			return blockParsed
		}

	}

	// Find from other nodes

	quorumUrlsAndPubkeys := utils.GetQuorumUrlsAndPubkeys(epochHandler)

	var quorumUrls []string

	for _, quorumMember := range quorumUrlsAndPubkeys {

		quorumUrls = append(quorumUrls, quorumMember.Url)

	}

	allKnownNodes := append(quorumUrls, globals.CONFIGURATION.BootstrapNodes...)

	resultChan := make(chan *Block, len(allKnownNodes))
	var wg sync.WaitGroup

	for _, node := range allKnownNodes {

		if node == globals.CONFIGURATION.MyHostname {
			continue
		}

		wg.Add(1)
		go func(endpoint string) {

			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			url := endpoint + "/block/" + blockID
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				return
			}
			defer resp.Body.Close()

			var block Block

			if err := json.NewDecoder(resp.Body).Decode(&block); err == nil {
				resultChan <- &block
			}

		}(node)

	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for block := range resultChan {
		if block != nil {
			return block
		}
	}

	return nil
}
