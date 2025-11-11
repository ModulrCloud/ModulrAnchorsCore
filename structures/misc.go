package structures

type ResponseStatus struct {
	Status string
}

type QuorumMemberData struct {
	PubKey, Url string
}

type FirstBlockResult struct {
	FirstBlockCreator, FirstBlockHash string
}

type FirstBlockAssumption struct {
	IndexOfFirstBlockCreator int                         `json:"indexOfFirstBlockCreator"`
	AfpForSecondBlock        AggregatedFinalizationProof `json:"afpForSecondBlock"`
}

type ExecutionStatsPerLeaderSequence struct {
	Index          int
	Hash           string
	FirstBlockHash string
}

func NewExecutionStatsTemplate() ExecutionStatsPerLeaderSequence {

	return ExecutionStatsPerLeaderSequence{
		Index:          -1,
		Hash:           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		FirstBlockHash: "",
	}

}
