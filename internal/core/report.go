package core

// Report defines an output file produced by a workflow
type Report struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Type        string `yaml:"type"` // text, csv, json, etc.
	Description string `yaml:"description"`
	Optional    bool   `yaml:"optional,omitempty"`
}

// IsTextReport returns true if this is a text report
func (r *Report) IsTextReport() bool {
	return r.Type == "text" || r.Type == ""
}

// IsCSVReport returns true if this is a CSV report
func (r *Report) IsCSVReport() bool {
	return r.Type == "csv"
}

// IsJSONReport returns true if this is a JSON report
func (r *Report) IsJSONReport() bool {
	return r.Type == "json"
}
