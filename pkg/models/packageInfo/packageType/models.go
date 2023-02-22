package packageType

import (
	"database/sql/driver"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/iconInfo"
)

// PackageType is an enum indicating the type of the Package
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

func (s Type) String() string {
	switch s {
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

func (s Type) DBMap(value string) Type {
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

func (u *Type) Scan(value interface{}) error { *u = u.DBMap(value.(string)); return nil }

func (u Type) Value() (driver.Value, error) { return u.String(), nil }

type Info struct {
	PackageType    Type
	PackageSubType string
	Icon           iconInfo.Icon
	HasGrouping    bool
}

// fileTypeDict maps filetypes to PackageTypes.
var FileTypeToInfoDict = map[fileType.Type]Info{
	fileType.MEF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.EDF: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.TDMS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.OpenEphys: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Persyst: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
		HasGrouping:    true,
	},
	fileType.NeuroExplorer: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.MobergSeries: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.BFTS: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Nicolet: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.MEF3: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Feather: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.NEV: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Spike2: {
		PackageType:    TimeSeries,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.MINC: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.DICOM: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.NIFTI: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.ROI: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.SWC: {
		PackageType:    Unsupported,
		PackageSubType: "Morphology",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.ANALYZE: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.MGH: {
		PackageType:    MRI,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.JPEG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileType.PNG: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileType.TIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.OMETIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.BRUKERTIFF: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.CZI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.JPEG2000: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.LSM: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.NDPI: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.OIB: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.OIF: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.GIF: {
		PackageType:    Image,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileType.WEBM: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileType.MOV: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileType.AVI: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileType.MP4: {
		PackageType:    Video,
		PackageSubType: "Video",
		Icon:           iconInfo.Video,
	},
	fileType.CSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Tabular,
	},
	fileType.TSV: {
		PackageType:    CSV,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Tabular,
	},
	fileType.MSExcel: {
		PackageType:    Unsupported,
		PackageSubType: "MS Excel",
		Icon:           iconInfo.Excel,
	},
	fileType.Aperio: {
		PackageType:    Slide,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.MSWord: {
		PackageType:    MSWord,
		PackageSubType: "MS Word",
		Icon:           iconInfo.Word,
	},
	fileType.PDF: {
		PackageType:    PDF,
		PackageSubType: "PDF",
		Icon:           iconInfo.PDF,
	},
	fileType.Text: {
		PackageType:    Text,
		PackageSubType: "Text",
		Icon:           iconInfo.Text,
	},
	fileType.BFANNOT: {
		PackageType:    Unknown,
		PackageSubType: "Text",
		Icon:           iconInfo.Generic,
	},
	fileType.AdobeIllustrator: {
		PackageType:    Unsupported,
		PackageSubType: "Illustrator",
		Icon:           iconInfo.AdobeIllustrator,
	},
	fileType.AFNI: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.AFNIBRIK: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.Ansys: {
		PackageType:    Unsupported,
		PackageSubType: "Ansys",
		Icon:           iconInfo.Code,
	},
	fileType.BAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           iconInfo.Genomics,
	},
	fileType.CRAM: {
		PackageType:    Unsupported,
		PackageSubType: "Genomics",
		Icon:           iconInfo.Genomics,
	},
	fileType.BIODAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.BioPAC: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.COMSOL: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           iconInfo.Model,
	},
	fileType.CPlusPlus: {
		PackageType:    Unsupported,
		PackageSubType: "C++",
		Icon:           iconInfo.Code,
	},
	fileType.CSharp: {
		PackageType:    Unsupported,
		PackageSubType: "C#",
		Icon:           iconInfo.Code,
	},
	fileType.Data: {
		PackageType:    Unsupported,
		PackageSubType: "generic",
		Icon:           iconInfo.GenericData,
	},
	fileType.Docker: {
		PackageType:    Unsupported,
		PackageSubType: "Docker",
		Icon:           iconInfo.Docker,
	},
	fileType.EPS: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Image,
	},
	fileType.FCS: {
		PackageType:    Unsupported,
		PackageSubType: "Flow",
		Icon:           iconInfo.Flow,
	},
	fileType.FASTA: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Genomics,
	},
	fileType.FASTQ: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.Genomics,
	},
	fileType.FreesurferSurface: {
		PackageType:    Unsupported,
		PackageSubType: "3D Image",
		Icon:           iconInfo.ClinicalImageBrain,
	},
	fileType.HDF: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           iconInfo.HDF,
	},
	fileType.HTML: {
		PackageType:    Unsupported,
		PackageSubType: "HTML",
		Icon:           iconInfo.Code,
	},
	fileType.Imaris: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.Intan: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.IVCurveData: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.JAVA: {
		PackageType:    Unsupported,
		PackageSubType: "JAVA",
		Icon:           iconInfo.Code,
	},
	fileType.Javascript: {
		PackageType:    Unsupported,
		PackageSubType: "Javascript",
		Icon:           iconInfo.Code,
	},
	fileType.Json: {
		PackageType:    Unsupported,
		PackageSubType: "JSON",
		Icon:           iconInfo.JSON,
	},
	fileType.Jupyter: {
		PackageType:    Unsupported,
		PackageSubType: "Notebook",
		Icon:           iconInfo.Notebook,
	},
	fileType.LabChart: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Leica: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.MATLAB: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.Matlab,
	},
	fileType.MatlabFigure: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Matlab,
	},
	fileType.Markdown: {
		PackageType:    Unsupported,
		PackageSubType: "Markdown",
		Icon:           iconInfo.Code,
	},
	fileType.Minitab: {
		PackageType:    Unsupported,
		PackageSubType: "Generic",
		Icon:           iconInfo.GenericData,
	},
	fileType.Neuralynx: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.NeuroDataWithoutBorders: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.NWB,
	},
	fileType.Neuron: {
		PackageType:    Unsupported,
		PackageSubType: "Code",
		Icon:           iconInfo.Code,
	},
	fileType.NihonKoden: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Nikon: {
		PackageType:    Unsupported,
		PackageSubType: "Image",
		Icon:           iconInfo.Microscope,
	},
	fileType.PatchMaster: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.PClamp: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.Plexon: {
		PackageType:    Unsupported,
		PackageSubType: "Timeseries",
		Icon:           iconInfo.Timeseries,
	},
	fileType.PowerPoint: {
		PackageType:    Unsupported,
		PackageSubType: "MS Powerpoint",
		Icon:           iconInfo.PowerPoint,
	},
	fileType.Python: {
		PackageType:    Unsupported,
		PackageSubType: "Python",
		Icon:           iconInfo.Code,
	},
	fileType.R: {
		PackageType:    Unsupported,
		PackageSubType: "R",
		Icon:           iconInfo.Code,
	},
	fileType.RData: {
		PackageType:    Unsupported,
		PackageSubType: "Data Container",
		Icon:           iconInfo.RData,
	},
	fileType.Shell: {
		PackageType:    Unsupported,
		PackageSubType: "Shell",
		Icon:           iconInfo.Code,
	},
	fileType.SolidWorks: {
		PackageType:    Unsupported,
		PackageSubType: "Model",
		Icon:           iconInfo.Model,
	},
	fileType.VariantData: {
		PackageType:    Unsupported,
		PackageSubType: "Tabular",
		Icon:           iconInfo.GenomicsVariant,
	},
	fileType.XML: {
		PackageType:    Unsupported,
		PackageSubType: "XML",
		Icon:           iconInfo.XML,
	},
	fileType.YAML: {
		PackageType:    Unsupported,
		PackageSubType: "YAML",
		Icon:           iconInfo.Code,
	},
	fileType.ZIP: {
		PackageType:    ZIP,
		PackageSubType: "ZIP",
		Icon:           iconInfo.Zip,
	},
	fileType.HDF5: {
		PackageType:    HDF5,
		PackageSubType: "Data Container",
		Icon:           iconInfo.HDF,
	},
	fileType.GenericData: {
		PackageType:    Unsupported,
		PackageSubType: "Generic Data",
		Icon:           iconInfo.Generic,
	},
}
