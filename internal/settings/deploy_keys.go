package settings

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// DeployKey declares a public SSH key attached to the repository. Key is the
// stable identity; private-key material is never accepted.
type DeployKey struct {
	Key      string `json:"key" yaml:"key"`
	Title    string `json:"title,omitempty" yaml:"title,omitempty"`
	ReadOnly *bool  `json:"read_only,omitempty" yaml:"read_only,omitempty"`
	State    string `json:"state,omitempty" yaml:"state,omitempty"`
}

// ValidateDeployKeys validates the manifest declarations for deploy keys.
func ValidateDeployKeys(keys []DeployKey) error {
	seen := make(map[string]struct{}, len(keys))
	for i, key := range keys {
		identity, err := DeployKeyIdentity(key.Key)
		if err != nil {
			return fmt.Errorf("deploy_keys[%d].key %w", i, err)
		}
		if _, ok := seen[identity]; ok {
			return fmt.Errorf("deploy_keys[%d].key duplicates an earlier deploy key", i)
		}
		seen[identity] = struct{}{}
		switch key.State {
		case "", "present":
			if strings.TrimSpace(key.Title) == "" {
				return fmt.Errorf("deploy_keys[%d].title is required", i)
			}
			if key.ReadOnly == nil {
				return fmt.Errorf("deploy_keys[%d].read_only is required", i)
			}
		case "absent":
			if key.Title != "" || key.ReadOnly != nil {
				return fmt.Errorf("deploy_keys[%d] cannot include title or read_only when state is absent", i)
			}
		default:
			return fmt.Errorf("deploy_keys[%d].state must be present or absent", i)
		}
	}
	return nil
}

// DeployKeyIdentity returns the key type and base64-encoded public-key data,
// which together identify a deploy key regardless of its optional comment.
func DeployKeyIdentity(raw string) (string, error) {
	fields := strings.Fields(raw)
	if len(fields) < 2 {
		return "", fmt.Errorf("must be an SSH public key")
	}
	switch fields[0] {
	case "ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521",
		"sk-ssh-ed25519@openssh.com", "sk-ecdsa-sha2-nistp256@openssh.com":
	default:
		return "", fmt.Errorf("must use a supported SSH public-key type")
	}
	decoded, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil || len(decoded) == 0 {
		return "", fmt.Errorf("must contain valid base64 public-key data")
	}
	return fields[0] + " " + fields[1], nil
}
