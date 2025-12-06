package utils

import (
	"encoding/json"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/globals"
	"github.com/modulrcloud/modulr-anchors-core/structures"

	"github.com/syndtr/goleveldb/leveldb"
)

func OpenDb(dbName string) *leveldb.DB {

	db, err := leveldb.OpenFile(globals.CHAINDATA_PATH+"/DATABASES/"+dbName, nil)

	if err != nil {
		panic("Impossible to open db : " + dbName + " =>" + err.Error())
	}

	return db

}

func GetAnchorFromApprovementThreadState(anchorPubkey string) *structures.AnchorStorage {

	anchorStorageKey := anchorPubkey + "_ANCHOR_STORAGE"

	data, err := databases.APPROVEMENT_THREAD_METADATA.Get([]byte(anchorStorageKey), nil)

	if err != nil {
		return nil
	}

	var anchorStorage structures.AnchorStorage

	err = json.Unmarshal(data, &anchorStorage)

	if err != nil {
		return nil
	}

	return &anchorStorage

}
