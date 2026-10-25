package packageType

import (
	"testing"

	"github.com/pennsieve/pennsieve-go-core/pkg/models/fileInfo/fileType"
)

// As pennsieve-api's FileTypeInfo: MAT data files are HDF5 data containers,
// MATLAB scripts are unsupported code.
func TestMatlabPackageInfo(t *testing.T) {
	if got := FileTypeToInfoDict[fileType.MAT]; got.PackageType != HDF5 || got.PackageSubType != "Data Container" {
		t.Errorf("MAT info = %+v", got)
	}
	if got := FileTypeToInfoDict[fileType.MATLAB]; got.PackageType != Unsupported || got.PackageSubType != "Matlab" {
		t.Errorf("MATLAB info = %+v", got)
	}
}
