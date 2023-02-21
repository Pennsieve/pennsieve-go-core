package packageState

import "database/sql/driver"

// PackageState is an enum indicating the state of the Package
type State int64

const (
	Unavailable State = iota
	Uploaded
	Deleting
	Infected
	UploadFailed
	Processing
	Ready
	ProcessingFailed
)

func (s State) String() string {
	switch s {
	case Unavailable:
		return "UNAVAILABLE"
	case Uploaded:
		return "UPLOADED"
	case Deleting:
		return "DELETING"
	case Infected:
		return "INFECTED"
	case UploadFailed:
		return "UPLOAD_FAILED"
	case Processing:
		return "PROCESSING"
	case Ready:
		return "READY"
	case ProcessingFailed:
		return "PROCESSING_FAILED"
	}
	return "UNKNOWN"
}

func (s State) Dict(value string) State {
	switch value {
	case "UNAVAILABLE":
		return Unavailable
	case "UPLOADED":
		return Uploaded
	case "DELETING":
		return Deleting
	case "INFECTED":
		return Infected
	case "UPLOAD_FAILED":
		return UploadFailed
	case "PROCESSING":
		return Processing
	case "READY":
		return Ready
	case "PROCESSING_FAILED":
		return ProcessingFailed
	}
	return Unavailable
}

func (u *State) Scan(value interface{}) error { *u = u.Dict(value.(string)); return nil }
func (u State) Value() (driver.Value, error)  { return u.String(), nil }
