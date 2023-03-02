package packageType

import (
	"database/sql/driver"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/iconInfo"
)

// Type is an enum indicating the type of the Package
type Type int64

const (
	Image Type = iota
	MRI
	Slide
	ExternalFile
	MSWord
	PDF
	CSV
	Tabular
	TimeSeries
	Video
	Unknown
	Collection
	Text
	Unsupported
	HDF5
	ZIP
)

func (t Type) String() string {
	switch t {
	case Image:
		return "Image"
	case MRI:
		return "MRI"
	case Slide:
		return "Slide"
	case ExternalFile:
		return "ExternalFile"
	case MSWord:
		return "MSWord"
	case PDF:
		return "PDF"
	case CSV:
		return "CSV"
	case Tabular:
		return "Tabular"
	case TimeSeries:
		return "TimeSeries"
	case Video:
		return "Video"
	case Unknown:
		return "Unknown"
	case Collection:
		return "Collection"
	case Text:
		return "Text"
	case Unsupported:
		return "Unsupported"
	case HDF5:
		return "HDF5"
	case ZIP:
		return "ZIP"
	}
	return "Unknown"
}

func (t Type) DBMap(value string) Type {
	switch value {
	case "Image":
		return Image
	case "MRI":
		return MRI
	case "Slide":
		return Slide
	case "ExternalFile":
		return ExternalFile
	case "MSWord":
		return MSWord
	case "PDF":
		return PDF
	case "CSV":
		return CSV
	case "Tabular":
		return Tabular
	case "TimeSeries":
		return TimeSeries
	case "Video":
		return Video
	case "Unknown":
		return Unknown
	case "Collection":
		return Collection
	case "Text":
		return Text
	case "Unsupported":
		return Unsupported
	case "HDF5":
		return HDF5
	case "ZIP":
		return ZIP
	}
	return Unknown
}

func (t *Type) Scan(value interface{}) error { *t = t.DBMap(value.(string)); return nil }

func (t Type) Value() (driver.Value, error) { return t.String(), nil }

type Info struct {
	PackageType    Type
	PackageSubType string
	Icon           iconInfo.Icon
	HasGrouping    bool
	HasWorkflow    bool
}

// FileTypeToInfoDict maps filetypes to PackageTypes.
var FileTypeToInfoDict = map[fileType.Type]Info{
	fileType.MEF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.EDF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.TDMS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.OpenEphys: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.Persyst: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasGrouping:    true,
		HasWorkflow:    true,
	},
	fileType.NeuroExplorer: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.MobergSeries: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.BFTS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.Nicolet: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.MEF3: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.Feather: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.NEV: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.Spike2: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    true,
	},
	fileType.MINC: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    true,
	},
	fileType.DICOM: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    true,
	},
	fileType.NIFTI: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    true,
	},
	fileType.ROI: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    false,
	},
	fileType.SWC: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    false,
	},
	fileType.ANALYZE: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    true,
	},
	fileType.MGH: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    true,
	},
	fileType.JPEG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
		HasWorkflow:    true,
	},
	fileType.PNG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
		HasWorkflow:    true,
	},
	fileType.TIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    true,
	},
	fileType.OMETIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.BRUKERTIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    true,
	},
	fileType.CZI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.JPEG2000: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    true,
	},
	fileType.LSM: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.NDPI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.OIB: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.OIF: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.GIF: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
		HasWorkflow:    true,
	},
	fileType.WEBM: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
		HasWorkflow:    true,
	},
	fileType.MOV: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
		HasWorkflow:    true,
	},
	fileType.AVI: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
		HasWorkflow:    true,
	},
	fileType.MP4: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
		HasWorkflow:    true,
	},
	fileType.CSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Tabular,
		HasWorkflow:    false,
	},
	fileType.TSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Tabular,
		HasWorkflow:    false,
	},
	fileType.MSExcel: {
		PackageType:    Unsupported,
		PackageSubType: "MS Excel",
		Icon:           iconInfo.Excel,
		HasWorkflow:    false,
	},
	fileType.Aperio: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    true,
	},
	fileType.MSWord: {
		PackageType:    MSWord,
		PackageSubType: "MS Word",
		Icon:           iconInfo.Word,
		HasWorkflow:    false,
	},
	fileType.PDF: {
		PackageType:    PDF,
		PackageSubType: "PDF",
		Icon:           iconInfo.PDF,
		HasWorkflow:    false,
	},
	fileType.Text: {
		PackageType:    Text,
		PackageSubType: "Text",
		Icon:           iconInfo.Text,
		HasWorkflow:    false,
	},
	fileType.BFANNOT: {
		PackageType:    Unknown,
		PackageSubType: "Text",
		Icon:           iconInfo.Generic,
		HasWorkflow:    false,
	},
	fileType.AdobeIllustrator: {
		PackageType:    Unsupported,
		PackageSubType: "Illustrator",
		Icon:           iconInfo.AdobeIllustrator,
		HasWorkflow:    false,
	},
	fileType.AFNI: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    false,
	},
	fileType.AFNIBRIK: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    false,
	},
	fileType.Ansys: {
		PackageType:    Unsupported,
		PackageSubType: "Ansys",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.BAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           iconInfo.Genomics,
		HasWorkflow:    false,
	},
	fileType.CRAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           iconInfo.Genomics,
		HasWorkflow:    false,
	},
	fileType.BIODAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.BioPAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.COMSOL: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           iconInfo.Model,
		HasWorkflow:    false,
	},
	fileType.CPlusPlus: {
		PackageType:    Unsupported,
		PackageSubType: "C++",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.CSharp: {
		PackageType:    Unsupported,
		PackageSubType: "C#",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.Data: {
		PackageType:    Unsupported,
		PackageSubType: "generic",
		Icon:           iconInfo.GenericData,
		HasWorkflow:    false,
	},
	fileType.Docker: {
		PackageType:    Unsupported,
		PackageSubType: "Docker",
		Icon:           iconInfo.Docker,
		HasWorkflow:    false,
	},
	fileType.EPS: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
		HasWorkflow:    false,
	},
	fileType.FCS: {
		PackageType:    Unsupported,
		PackageSubType: "Flow",
		Icon:           iconInfo.Flow,
		HasWorkflow:    false,
	},
	fileType.FASTA: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Genomics,
		HasWorkflow:    false,
	},
	fileType.FASTQ: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Genomics,
		HasWorkflow:    false,
	},
	fileType.FreesurferSurface: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
		HasWorkflow:    false,
	},
	fileType.HDF: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           iconInfo.HDF,
		HasWorkflow:    false,
	},
	fileType.HTML: {
		PackageType:    Unsupported,
		PackageSubType: "HTML",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.Imaris: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.Intan: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.IVCurveData: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.JAVA: {
		PackageType:    Unsupported,
		PackageSubType: "JAVA",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.Javascript: {
		PackageType:    Unsupported,
		PackageSubType: "Javascript",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.Json: {
		PackageType:    Unsupported,
		PackageSubType: "JSON",
		Icon:           iconInfo.JSON,
		HasWorkflow:    false,
	},
	fileType.Jupyter: {
		PackageType:    Unsupported,
		PackageSubType: "Notebook",
		Icon:           iconInfo.Notebook,
		HasWorkflow:    false,
	},
	fileType.LabChart: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.Leica: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.MATLAB: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.Matlab,
		HasWorkflow:    false,
	},
	fileType.MatlabFigure: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Matlab,
		HasWorkflow:    false,
	},
	fileType.Markdown: {
		PackageType:    Unsupported,
		PackageSubType: "Markdown",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.Minitab: {
		PackageType:    Unsupported,
		PackageSubType: "Generic",
		Icon:           iconInfo.GenericData,
		HasWorkflow:    false,
	},
	fileType.Neuralynx: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.NeuroDataWithoutBorders: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.NWB,
		HasWorkflow:    false,
	},
	fileType.Neuron: {
		PackageType:    Unsupported,
		PackageSubType: "Code",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.NihonKoden: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.Nikon: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
		HasWorkflow:    false,
	},
	fileType.PatchMaster: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.PClamp: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.Plexon: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasWorkflow:    false,
	},
	fileType.PowerPoint: {
		PackageType:    Unsupported,
		PackageSubType: "MS Powerpoint",
		Icon:           iconInfo.PowerPoint,
		HasWorkflow:    false,
	},
	fileType.Python: {
		PackageType:    Unsupported,
		PackageSubType: "Python",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.R: {
		PackageType:    Unsupported,
		PackageSubType: "R",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.RData: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           iconInfo.RData,
		HasWorkflow:    false,
	},
	fileType.Shell: {
		PackageType:    Unsupported,
		PackageSubType: "Shell",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.SolidWorks: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           iconInfo.Model,
		HasWorkflow:    false,
	},
	fileType.VariantData: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.GenomicsVariant,
		HasWorkflow:    false,
	},
	fileType.XML: {
		PackageType:    Unsupported,
		PackageSubType: "XML",
		Icon:           iconInfo.XML,
		HasWorkflow:    false,
	},
	fileType.YAML: {
		PackageType:    Unsupported,
		PackageSubType: "YAML",
		Icon:           iconInfo.Code,
		HasWorkflow:    false,
	},
	fileType.ZIP: {
		PackageType:    ZIP,
		PackageSubType: "ZIP",
		Icon:           iconInfo.Zip,
		HasWorkflow:    false,
	},
	fileType.HDF5: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.HDF,
		HasWorkflow:    false,
	},
	fileType.GenericData: {
		PackageType:    Unsupported,
		PackageSubType: "Generic Data",
		Icon:           iconInfo.Generic,
		HasWorkflow:    false,
	},
}
