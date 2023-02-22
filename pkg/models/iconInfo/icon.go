package iconInfo

type Icon int64

const (
	AdobeIllustrator Icon = iota
	ClinicalImageBrain
	Code
	Docker
	Excel
	Flow
	Generic
	GenericData
	Genomics
	GenomicsVariant
	HDF
	Image
	JSON
	Matlab
	Microscope
	Model
	Notebook
	NWB
	PDF
	PowerPoint
	RData
	Tabular
	Text
	Timeseries
	Video
	Word
	XML
	Zip
)

func (s Icon) String() string {
	switch s {
	case AdobeIllustrator:
		return "AdobeIllustrator"
	case ClinicalImageBrain:
		return "ClinicalImageBrain"
	case Code:
		return "Code"
	case Docker:
		return "Docker"
	case Excel:
		return "Excel"
	case Flow:
		return "Flow"
	case Generic:
		return "Generic"
	case GenericData:
		return "GenericData"
	case Genomics:
		return "Genomics"
	case GenomicsVariant:
		return "GenomicsVariant"
	case HDF:
		return "HDF"
	case Image:
		return "Image"
	case JSON:
		return "JSON"
	case Matlab:
		return "Matlab"
	case Microscope:
		return "Microscope"
	case Model:
		return "Model"
	case Notebook:
		return "Notebook"
	case NWB:
		return "NWB"
	case PDF:
		return "PDF"
	case PowerPoint:
		return "PowerPoint"
	case RData:
		return "RData"
	case Tabular:
		return "Tabular"
	case Text:
		return "Text"
	case Timeseries:
		return "Timeseries"
	case Video:
		return "Video"
	case Word:
		return "Word"
	case XML:
		return "XML"
	case Zip:
		return "Zip"
	}
	return "Generic"
}
