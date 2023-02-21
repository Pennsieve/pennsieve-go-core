package fileType

// FileType is an enum indicating the type of the File
type Type int64

const (
	GenericData Type = iota
	PDF
	MEF
	EDF
	TDMS
	OpenEphys
	Persyst
	DICOM
	NIFTI
	PNG
	CZI
	Aperio
	Json
	CSV
	TSV
	Text
	XML
	HTML
	MSExcel
	MSWord
	MP4
	WEBM
	OGG
	MOV
	JPEG
	JPEG2000
	LSM
	NDPI
	OIB
	OIF
	ROI
	SWC
	CRAM
	MGH
	AVI
	MATLAB
	HDF5
	TIFF
	OMETIFF
	BRUKERTIFF
	GIF
	ANALYZE
	NeuroExplorer
	MINC
	MobergSeries
	BFANNOT
	BFTS
	Nicolet
	MEF3
	Feather
	NEV
	Spike2
	AdobeIllustrator
	AFNI
	AFNIBRIK
	Ansys
	BAM
	BIODAC
	BioPAC
	COMSOL
	CPlusPlus
	CSharp
	Data
	Docker
	EPS
	FCS
	FASTA
	FASTQ
	FreesurferSurface
	HDF
	Imaris
	Intan
	IVCurveData
	JAVA
	Javascript
	Jupyter
	LabChart
	Leica
	MatlabFigure
	Markdown
	Minitab
	Neuralynx
	NeuroDataWithoutBorders
	Neuron
	NihonKoden
	Nikon
	PatchMaster
	PClamp
	Plexon
	PowerPoint
	Python
	R
	RData
	Shell
	SolidWorks
	VariantData
	YAML
	ZIP
)

//String returns the string representation of the File.Type
func (s Type) String() string {
	switch s {
	case PDF:
		return "PDF"
	case MEF:
		return "MEF"
	case EDF:
		return "EDF"
	case TDMS:
		return "TDMS"
	case OpenEphys:
		return "OpenEphys"
	case Persyst:
		return "Persyst"
	case DICOM:
		return "DICOM"
	case NIFTI:
		return "NIFTI"
	case PNG:
		return "PNG"
	case CZI:
		return "CZI"
	case Aperio:
		return "Aperio"
	case Json:
		return "Json"
	case CSV:
		return "CSV"
	case TSV:
		return "TSV"
	case Text:
		return "Text"
	case XML:
		return "XML"
	case HTML:
		return "HTML"
	case MSExcel:
		return "MSExcel"
	case MSWord:
		return "MSWord"
	case MP4:
		return "MP4"
	case WEBM:
		return "WEBM"
	case OGG:
		return "OGG"
	case MOV:
		return "MOV"
	case JPEG:
		return "JPEG"
	case JPEG2000:
		return "JPEG2000"
	case LSM:
		return "LSM"
	case NDPI:
		return "NDPI"
	case OIB:
		return "OIB"
	case OIF:
		return "OIF"
	case ROI:
		return "ROI"
	case SWC:
		return "SWC"
	case CRAM:
		return "CRAM"
	case MGH:
		return "MGH"
	case AVI:
		return "AVI"
	case MATLAB:
		return "MATLAB"
	case HDF5:
		return "HDF5"
	case TIFF:
		return "TIFF"
	case OMETIFF:
		return "OMETIFF"
	case BRUKERTIFF:
		return "BRUKERTIFF"
	case GIF:
		return "GIF"
	case ANALYZE:
		return "ANALYZE"
	case NeuroExplorer:
		return "NeuroExplorer"
	case MINC:
		return "MINC"
	case MobergSeries:
		return "MobergSeries"
	case GenericData:
		return "GenericData"
	case BFANNOT:
		return "BFANNOT"
	case BFTS:
		return "BFTS"
	case Nicolet:
		return "Nicolet"
	case MEF3:
		return "MEF3"
	case Feather:
		return "Feather"
	case NEV:
		return "NEV"
	case Spike2:
		return "Spike2"
	case AdobeIllustrator:
		return "AdobeIllustrator"
	case AFNI:
		return "AFNI"
	case AFNIBRIK:
		return "AFNIBRIK"
	case Ansys:
		return "Ansys"
	case BAM:
		return "BAM"
	case BIODAC:
		return "BIODAC"
	case BioPAC:
		return "BioPAC"
	case COMSOL:
		return "COMSOL"
	case CPlusPlus:
		return "CPlusPlus"
	case CSharp:
		return "CSharp"
	case Data:
		return "Data"
	case Docker:
		return "Docker"
	case EPS:
		return "EPS"
	case FCS:
		return "FCS"
	case FASTA:
		return "FASTA"
	case FASTQ:
		return "FASTQ"
	case FreesurferSurface:
		return "FreesurferSurface"
	case HDF:
		return "HDF"
	case Imaris:
		return "Imaris"
	case Intan:
		return "Intan"
	case IVCurveData:
		return "IVCurveData"
	case JAVA:
		return "JAVA"
	case Javascript:
		return "Javascript"
	case Jupyter:
		return "Jupyter"
	case LabChart:
		return "LabChart"
	case Leica:
		return "Leica"
	case MatlabFigure:
		return "MatlabFigure"
	case Markdown:
		return "Markdown"
	case Minitab:
		return "Minitab"
	case Neuralynx:
		return "Neuralynx"
	case NeuroDataWithoutBorders:
		return "NeuroDataWithoutBorders"
	case Neuron:
		return "Neuron"
	case NihonKoden:
		return "NihonKoden"
	case Nikon:
		return "Nikon"
	case PatchMaster:
		return "PatchMaster"
	case PClamp:
		return "PClamp"
	case Plexon:
		return "Plexon"
	case PowerPoint:
		return "PowerPoint"
	case Python:
		return "Python"
	case R:
		return "R"
	case RData:
		return "RData"
	case Shell:
		return "Shell"
	case SolidWorks:
		return "SolidWorks"
	case VariantData:
		return "VariantData"
	case YAML:
		return "YAML"
	case ZIP:
		return "ZIP"
	}
	return "GenericData"
}

// Dict maps string to FileType object.
var Dict = map[string]Type{
	"PDF":                     PDF,
	"MEF":                     MEF,
	"EDF":                     EDF,
	"TDMS":                    TDMS,
	"OpenEphys":               OpenEphys,
	"Persyst":                 Persyst,
	"DICOM":                   DICOM,
	"NIFTI":                   NIFTI,
	"PNG":                     PNG,
	"CZI":                     CZI,
	"Aperio":                  Aperio,
	"Json":                    Json,
	"CSV":                     CSV,
	"TSV":                     TSV,
	"Text":                    Text,
	"XML":                     XML,
	"HTML":                    HTML,
	"MSExcel":                 MSExcel,
	"MSWord":                  MSWord,
	"MP4":                     MP4,
	"WEBM":                    WEBM,
	"OGG":                     OGG,
	"MOV":                     MOV,
	"JPEG":                    JPEG,
	"JPEG2000":                JPEG2000,
	"LSM":                     LSM,
	"NDPI":                    NDPI,
	"OIB":                     OIB,
	"OIF":                     OIF,
	"ROI":                     ROI,
	"SWC":                     SWC,
	"CRAM":                    CRAM,
	"MGH":                     MGH,
	"AVI":                     AVI,
	"MATLAB":                  MATLAB,
	"HDF5":                    HDF5,
	"TIFF":                    TIFF,
	"OMETIFF":                 OMETIFF,
	"BRUKERTIFF":              BRUKERTIFF,
	"GIF":                     GIF,
	"ANALYZE":                 ANALYZE,
	"NeuroExplorer":           NeuroExplorer,
	"MINC":                    MINC,
	"MobergSeries":            MobergSeries,
	"GenericData":             GenericData,
	"BFANNOT":                 BFANNOT,
	"BFTS":                    BFTS,
	"Nicolet":                 Nicolet,
	"MEF3":                    MEF3,
	"Feather":                 Feather,
	"NEV":                     NEV,
	"Spike2":                  Spike2,
	"AdobeIllustrator":        AdobeIllustrator,
	"AFNI":                    AFNI,
	"AFNIBRIK":                AFNIBRIK,
	"Ansys":                   Ansys,
	"BAM":                     BAM,
	"BIODAC":                  BIODAC,
	"BioPAC":                  BioPAC,
	"COMSOL":                  COMSOL,
	"CPlusPlus":               CPlusPlus,
	"CSharp":                  CSharp,
	"Data":                    Data,
	"Docker":                  Docker,
	"EPS":                     EPS,
	"FCS":                     FCS,
	"FASTA":                   FASTA,
	"FASTQ":                   FASTQ,
	"FreesurferSurface":       FreesurferSurface,
	"HDF":                     HDF5,
	"Imaris":                  Imaris,
	"Intan":                   Intan,
	"IVCurveData":             IVCurveData,
	"JAVA":                    JAVA,
	"Javascript":              Javascript,
	"Jupyter":                 Jupyter,
	"LabChart":                LabChart,
	"Leica":                   Leica,
	"MatlabFigure":            MatlabFigure,
	"Markdown":                Markdown,
	"Minitab":                 Minitab,
	"Neuralynx":               Neuralynx,
	"NeuroDataWithoutBorders": NeuroDataWithoutBorders,
	"Neuron":                  Neuron,
	"NihonKoden":              NihonKoden,
	"Nikon":                   Nikon,
	"PatchMaster":             PatchMaster,
	"PClamp":                  PClamp,
	"Plexon":                  Plexon,
	"PowerPoint":              PowerPoint,
	"Python":                  Python,
	"R":                       R,
	"RData":                   RData,
	"Shell":                   Shell,
	"SolidWorks":              SolidWorks,
	"VariantData":             VariantData,
	"YAML":                    YAML,
	"ZIP":                     ZIP,
}

// ExtensionToTypeDict maps file extensions to FileType object.
var ExtensionToTypeDict = map[string]Type{
	"bfannot":       BFANNOT,
	"bfts":          BFTS,
	"png":           PNG,
	"jpg":           JPEG,
	"jpeg":          JPEG,
	"jp2":           JPEG2000,
	"jpx":           JPEG2000,
	"lsm":           LSM,
	"ndpi":          NDPI,
	"oib":           OIB,
	"oif":           OIF,
	"ome.tiff":      OMETIFF,
	"ome.tif":       OMETIFF,
	"ome.tf2":       OMETIFF,
	"ome.tf8":       OMETIFF,
	"ome.btf":       OMETIFF,
	"brukertiff.gz": BRUKERTIFF,
	"tiff":          TIFF,
	"tif":           TIFF,
	"gif":           GIF,
	"ai":            AdobeIllustrator,
	"svg":           AdobeIllustrator,
	"nd2":           Nikon,
	"lif":           Leica,
	"ims":           Imaris,
	"txt":           Text,
	"text":          Text,
	"rtf":           Text,
	"html":          HTML,
	"htm":           HTML,
	"csv":           CSV,
	"pdf":           PDF,
	"doc":           MSWord,
	"docx":          MSWord,
	"json":          Json,
	"xls":           MSExcel,
	"xlsx":          MSExcel,
	"xml":           XML,
	"tsv":           TSV,
	"ppt":           PowerPoint,
	"pptx":          PowerPoint,
	"mat":           MATLAB,
	"mex":           MATLAB,
	"m":             MATLAB,
	"fig":           MatlabFigure,
	"mef":           MEF,
	"mefd.gz":       MEF3,
	"edf":           EDF,
	"tdm":           TDMS,
	"tdms":          TDMS,
	"lay":           Persyst,
	"dat":           Data,
	"nex":           NeuroExplorer,
	"nex5":          NeuroExplorer,
	"smr":           Spike2,
	".eeg":          NihonKoden,
	"plx":           Plexon,
	"pl2":           Plexon,
	"e":             Nicolet,
	"continuous":    OpenEphys,
	"spikes":        OpenEphys,
	"events":        OpenEphys,
	"openephys":     OpenEphys,
	"nev":           NEV,
	"ns1":           NEV,
	"ns2":           NEV,
	"ns3":           NEV,
	"ns4":           NEV,
	"ns5":           NEV,
	"ns6":           NEV,
	"nf3":           NEV,
	"moberg.gz":     MobergSeries,
	"feather":       Feather,
	"tab":           BIODAC,
	"acq":           BioPAC,
	"rhd":           Intan,
	"ibw":           IVCurveData,
	"adicht":        LabChart,
	"adidat":        LabChart,
	"ncs":           Neuralynx,
	"pgf":           PatchMaster,
	"pul":           PatchMaster,
	"abf":           PClamp,
	"dcm":           DICOM,
	"dicom":         DICOM,
	"nii":           NIFTI,
	"nii.gz":        NIFTI,
	"nifti":         NIFTI,
	"roi":           ROI,
	"swc":           SWC,
	"mgh":           MGH,
	"mgz":           MGH,
	"mgh.gz":        MGH,
	"mnc":           MINC,
	"img":           ANALYZE,
	"hdr":           ANALYZE,
	"afni":          AFNI,
	"brik":          AFNIBRIK,
	"head":          AFNIBRIK,
	"lh":            FreesurferSurface,
	"rh":            FreesurferSurface,
	"curv":          FreesurferSurface,
	"eps":           EPS,
	"ps":            EPS,
	"svs":           Aperio,
	"czi":           CZI,
	"mov":           MOV,
	"mp4":           MP4,
	"ogg":           OGG,
	"ogv":           OGG,
	"webm":          WEBM,
	"avi":           AVI,
	"mph":           COMSOL,
	"sldasm":        SolidWorks,
	"slddrw":        SolidWorks,
	"hdf":           HDF,
	"hdf4":          HDF,
	"hdf5":          HDF5,
	"h5":            HDF5,
	"h4":            HDF,
	"he2":           HDF,
	"he5":           HDF,
	"mpj":           Minitab,
	"mtw":           Minitab,
	"mgf":           Minitab,
	"nwb":           NeuroDataWithoutBorders,
	"rdata":         RData,
	"zip":           ZIP,
	"tar":           ZIP,
	"tar.gz":        ZIP,
	"gz":            ZIP,
	"fcs":           FCS,
	"bam":           BAM,
	"bcl":           BAM,
	"bcl.gz":        BAM,
	"fasta":         FASTA,
	"fasta.gz":      FASTA,
	"fastq":         FASTQ,
	"fastq.gz":      FASTQ,
	"vcf":           VariantData,
	"cram":          CRAM,
	"cs":            CSharp,
	"aedt":          Ansys,
	"cpp":           CPlusPlus,
	"js":            Javascript,
	"md":            Markdown,
	"hoc":           Neuron,
	"mod":           Neuron,
	"py":            Python,
	"r":             R,
	"sh":            Shell,
	"tcsh":          Shell,
	"bash":          Shell,
	"zsh":           Shell,
	"yaml":          YAML,
	"yml":           YAML,
	"java":          JAVA,
	"data":          Data,
	"bin":           Data,
	"raw":           Data,
	"Dockerfile":    Docker,
	"ipynb":         Jupyter,
}
