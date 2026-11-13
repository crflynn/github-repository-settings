package settings

// Environment declares one deployment environment. Omitted fields preserve their current value.
type Environment struct {
	Name                   string                  `json:"-" yaml:"name"`
	State                  string                  `json:"-" yaml:"state,omitempty"`
	WaitTimer              *int                    `json:"wait_timer,omitempty" yaml:"wait_timer,omitempty"`
	PreventSelfReview      *bool                   `json:"prevent_self_review,omitempty" yaml:"prevent_self_review,omitempty"`
	CanAdminsBypass        *bool                   `json:"can_admins_bypass,omitempty" yaml:"can_admins_bypass,omitempty"`
	Reviewers              *[]EnvironmentReviewer  `json:"reviewers,omitempty" yaml:"reviewers,omitempty"`
	DeploymentBranchPolicy *DeploymentBranchPolicy `json:"deployment_branch_policy,omitempty" yaml:"deployment_branch_policy,omitempty"`
}

type EnvironmentReviewer struct {
	Type string `json:"type" yaml:"type"`
	ID   int64  `json:"id" yaml:"id"`
}

type DeploymentBranchPolicy struct {
	ProtectedBranches    bool `json:"protected_branches" yaml:"protected_branches"`
	CustomBranchPolicies bool `json:"custom_branch_policies" yaml:"custom_branch_policies"`
}
