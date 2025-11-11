package utils

import (
	"encoding/json"

	"github.com/ModulrCloud/ModulrAnchorsCore/databases"
	"github.com/ModulrCloud/ModulrAnchorsCore/globals"
	"github.com/ModulrCloud/ModulrAnchorsCore/handlers"
	"github.com/ModulrCloud/ModulrAnchorsCore/structures"

	"github.com/syndtr/goleveldb/leveldb"
)

func OpenDb(dbName string) *leveldb.DB {

	db, err := leveldb.OpenFile(globals.CHAINDATA_PATH+"/DATABASES/"+dbName, nil)
	if err != nil {
		panic("Impossible to open db : " + dbName + " =>" + err.Error())
	}
	return db
}

func GetValidatorFromApprovementThreadState(validatorPubkey string) *structures.ValidatorStorage {

	validatorStorageKey := validatorPubkey + "_VALIDATOR_STORAGE"

	if val, ok := handlers.APPROVEMENT_THREAD_METADATA.Handler.ValidatorsStoragesCache[validatorStorageKey]; ok {
		return val
	}

	data, err := databases.APPROVEMENT_THREAD_METADATA.Get([]byte(validatorStorageKey), nil)

	if err != nil {
		return nil
	}

	var validatorStorage structures.ValidatorStorage

	err = json.Unmarshal(data, &validatorStorage)

	if err != nil {
		return nil
	}

	handlers.APPROVEMENT_THREAD_METADATA.Handler.ValidatorsStoragesCache[validatorStorageKey] = &validatorStorage

	return &validatorStorage

}
