package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

// Client wraps go-github with repository-scoped typed reconciliation methods.
type Client struct {
	api    *gh.Client
	owner  string
	repo   string
	dryRun *dryRunTransport
}

// NewClient constructs an authenticated go-github client for owner/name.
func NewClient(apiURL, token, repository string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}
	owner, repo, ok := strings.Cut(repository, "/")
	if !ok || !validRepositorySegment(owner) || !validRepositorySegment(repo) || strings.Contains(repo, "/") {
		return nil, fmt.Errorf("repository must be valid owner/name path segments, got %q", repository)
	}
	if apiURL == "" {
		apiURL = "https://api.github.com/"
	}
	parsed, err := url.Parse(apiURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("invalid GitHub API URL %q", apiURL)
	}
	baseURL := strings.TrimRight(apiURL, "/") + "/"
	options := []gh.ClientOptionsFunc{
		gh.WithAuthToken(token),
		gh.WithTimeout(30 * time.Second),
		gh.WithURLs(&baseURL, nil),
	}
	if httpClient != nil {
		options = append(options, gh.WithHTTPClient(httpClient))
	}
	api, err := gh.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("create GitHub API client: %w", err)
	}
	return &Client{api: api, owner: owner, repo: repo}, nil
}

// NewDryRunClient constructs a client that forwards reads and blocks all other
// API requests while recording their method and path.
func NewDryRunClient(apiURL, token, repository string) (*Client, error) {
	transport := &dryRunTransport{base: http.DefaultTransport}
	client, err := NewClient(apiURL, token, repository, &http.Client{Transport: transport})
	if err != nil {
		return nil, err
	}
	client.dryRun = transport
	return client, nil
}

// PlannedOperations returns the write requests intercepted during Apply.
func (c *Client) PlannedOperations() []PlannedOperation {
	if c.dryRun == nil {
		return nil
	}
	return c.dryRun.Operations()
}

// Apply reconciles only the components present in the manifest. Omitted components and resources are unchanged.
func (c *Client) Apply(ctx context.Context, desired settings.Settings) error {
	if err := desired.Validate(); err != nil {
		return err
	}
	if desired.Repository != nil {
		if err := c.applyRepository(ctx, *desired.Repository); err != nil {
			return fmt.Errorf("repository: %w", err)
		}
	}
	if desired.Actions != nil {
		if err := c.applyActions(ctx, *desired.Actions); err != nil {
			return fmt.Errorf("actions: %w", err)
		}
	}
	for _, collaborator := range desired.Collaborators {
		if err := c.applyCollaborator(ctx, collaborator); err != nil {
			return fmt.Errorf("collaborator %q: %w", collaborator.Username, err)
		}
	}
	for _, team := range desired.Teams {
		if err := c.applyTeamRepository(ctx, team); err != nil {
			return fmt.Errorf("team %q: %w", team.Slug, err)
		}
	}
	for _, environment := range desired.Environments {
		if err := c.applyEnvironment(ctx, environment); err != nil {
			return fmt.Errorf("environment %q: %w", environment.Name, err)
		}
	}
	if len(desired.Rulesets) > 0 {
		if err := c.applyRulesets(ctx, desired.Rulesets); err != nil {
			return fmt.Errorf("rulesets: %w", err)
		}
	}
	if desired.Pages != nil {
		if err := c.applyPages(ctx, *desired.Pages); err != nil {
			return fmt.Errorf("pages: %w", err)
		}
	}
	if len(desired.Autolinks) > 0 {
		if err := c.applyAutolinks(ctx, desired.Autolinks); err != nil {
			return fmt.Errorf("autolinks: %w", err)
		}
	}
	if len(desired.DeployKeys) > 0 {
		if err := c.applyDeployKeys(ctx, desired.DeployKeys); err != nil {
			return fmt.Errorf("deploy_keys: %w", err)
		}
	}
	if desired.InteractionLimits != nil {
		if err := c.applyInteractionLimits(ctx, *desired.InteractionLimits); err != nil {
			return fmt.Errorf("interaction_limits: %w", err)
		}
	}
	if len(desired.CustomProperties) > 0 {
		if err := c.applyCustomProperties(ctx, desired.CustomProperties); err != nil {
			return fmt.Errorf("custom_properties: %w", err)
		}
	}
	if len(desired.Webhooks) > 0 {
		if err := c.applyWebhooks(ctx, desired.Webhooks); err != nil {
			return fmt.Errorf("webhooks: %w", err)
		}
	}
	return nil
}

func isNotFound(resp *gh.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusNotFound
}

func validRepositorySegment(segment string) bool {
	if segment == "" || segment == "." || segment == ".." {
		return false
	}
	for _, char := range segment {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '_' && char != '.' {
			return false
		}
	}
	return true
}
