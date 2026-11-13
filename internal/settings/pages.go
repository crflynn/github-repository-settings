package settings

// Pages configures the repository's GitHub Pages site.
type Pages struct {
	Enabled       *bool        `json:"-" yaml:"enabled"`
	BuildType     string       `json:"build_type,omitempty" yaml:"build_type,omitempty"`
	Source        *PagesSource `json:"source,omitempty" yaml:"source,omitempty"`
	CNAME         *string      `json:"cname,omitempty" yaml:"cname,omitempty"`
	Public        *bool        `json:"public,omitempty" yaml:"public,omitempty"`
	HTTPSEnforced *bool        `json:"https_enforced,omitempty" yaml:"https_enforced,omitempty"`
}

type PagesSource struct {
	Branch string `json:"branch" yaml:"branch"`
	Path   string `json:"path" yaml:"path"`
}
