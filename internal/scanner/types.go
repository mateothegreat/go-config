package scanner

// StructInfo contains metadata about a struct to be validated
type StructInfo struct {
	Name         string
	PackageName  string
	Fields       []FieldInfo
	FilePath     string
	ImportPaths  []string
	Comments     string
}

// FieldInfo contains metadata about a struct field
type FieldInfo struct {
	Name          string
	Type          string
	Tag           string
	ValidateRules map[string]string
	JSONTag       string
	YAMLTag       string
	ConfigTag     string
	Required      bool
	Comments      string
}