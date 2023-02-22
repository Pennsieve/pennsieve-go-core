package uploadState

type UploadedState int64

const (
	Scanning UploadedState = iota
	Pending
	Uploaded
)

func (u UploadedState) String() string {
	switch u {
	case Scanning:
		return "SCANNING"
	case Pending:
		return "PENDING"
	case Uploaded:
		return "UPLOADED"
	default:
		return "UPLOADED"
	}
}

var Dict = map[string]UploadedState{
	"SCANNING": Scanning,
	"PENDING":  Pending,
	"UPLOADED": Uploaded,
}
