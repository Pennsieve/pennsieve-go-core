package processingState

type ProcessingState int64

const (
	Unprocessed ProcessingState = iota
	Processed
	NotProcessable
)

var Dict = map[string]ProcessingState{
	"unprocessed":     Unprocessed,
	"processed":       Processed,
	"not_processable": NotProcessable,
}

func (u ProcessingState) String() string {
	switch u {
	case Unprocessed:
		return "unprocessed"
	case Processed:
		return "processed"
	case NotProcessable:
		return "not_processable"
	default:
		return "unprocessed"
	}
}
