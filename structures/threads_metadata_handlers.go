package structures

type LogicalThread interface {
	GetCoreMajorVersion() int
	GetNetworkParams() NetworkParameters
	GetEpochHandler() EpochDataHandler
}

type ApprovementThreadMetadataHandler struct {
	CoreMajorVersion        int                          `json:"coreMajorVersion"`
	NetworkParameters       NetworkParameters            `json:"networkParameters"`
	EpochDataHandler        EpochDataHandler             `json:"epoch"`
	ValidatorsStoragesCache map[string]*ValidatorStorage `json:"-"`
}

func (handler *ApprovementThreadMetadataHandler) GetCoreMajorVersion() int {
	return handler.CoreMajorVersion
}

func (handler *ApprovementThreadMetadataHandler) GetNetworkParams() NetworkParameters {
	return handler.NetworkParameters
}

func (handler *ApprovementThreadMetadataHandler) GetEpochHandler() EpochDataHandler {
	return handler.EpochDataHandler
}

type GenerationThreadMetadataHandler struct {
	EpochFullId string `json:"epochFullId"`
	PrevHash    string `json:"prevHash"`
	NextIndex   int    `json:"nextIndex"`
}
