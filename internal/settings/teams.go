package settings

// TeamRepository grants an organization team access to this repository.
type TeamRepository struct {
	Slug       string `json:"-" yaml:"slug"`
	Permission string `json:"-" yaml:"permission"`
}
