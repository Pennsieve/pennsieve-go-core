package fileType

import "testing"

func TestMatlabExtensions(t *testing.T) {
	cases := map[string]Type{"mat": MAT, "mex": MAT, "m": MATLAB, "fig": MatlabFigure}
	for ext, want := range cases {
		if got := ExtensionToTypeDict[ext]; got != want {
			t.Errorf("ExtensionToTypeDict[%q] = %v, want %v", ext, got, want)
		}
	}
}

func TestMATRoundTrip(t *testing.T) {
	if MAT.String() != "MAT" {
		t.Fatalf("MAT.String() = %q", MAT.String())
	}
	if Dict["MAT"] != MAT {
		t.Fatalf(`Dict["MAT"] = %v`, Dict["MAT"])
	}
	// Appended: existing values keep their positions.
	if MAT <= ZIP {
		t.Fatalf("MAT (%d) must come after ZIP (%d)", MAT, ZIP)
	}
}

func TestZarrRoundTrip(t *testing.T) {
	for _, ft := range []Type{OMEZarr, TSZarr} {
		if got := Dict[ft.String()]; got != ft {
			t.Errorf("%s: Dict[String()] = %v", ft.String(), got)
		}
	}
	if OMEZarr.String() != "OMEZarr" || TSZarr.String() != "TSZarr" {
		t.Errorf("names: %s %s", OMEZarr.String(), TSZarr.String())
	}
}
