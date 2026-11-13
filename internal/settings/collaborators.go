package settings

// Collaborator declares one repository collaborator. Omitted collaborators are unchanged.
type Collaborator struct {
	Username   string `json:"-" yaml:"username"`
	Permission string `json:"-" yaml:"permission,omitempty"`
	State      string `json:"-" yaml:"state,omitempty"`
}
