package block_pack

import (
	"encoding/json"
)

type ExtraDataToBlock struct {
	Rest map[string]string `json:"rest"`
}

func (ed *ExtraDataToBlock) UnmarshalJSON(data []byte) error {

	type alias ExtraDataToBlock

	var aux alias

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Rest == nil {
		aux.Rest = make(map[string]string)
	}

	*ed = ExtraDataToBlock(aux)

	return nil

}
