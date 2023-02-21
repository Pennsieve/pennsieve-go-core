package packageInfo

type PackageAttribute struct {
	Key      string `json:"key"`
	Fixed    bool   `json:"fixed"`
	Value    string `json:"value"`
	Hidden   bool   `json:"hidden"`
	Category string `json:"category"`
	DataType string `json:"dataType"`
}
