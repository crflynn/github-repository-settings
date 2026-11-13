# GitHub Repository Settings

A locally maintained Go Docker Action for managing repository settings through GitHub's REST API. The internal client uses `github.com/google/go-github/v93` and maps a strict YAML schema to its typed API models.

## Local use

Build and run it from the repository being managed:

```sh
docker build -t github-repository-settings:local .
docker run --rm \
  --env INPUT_TOKEN \
  --env INPUT_SETTINGS-FILE=.github/repository-settings.yml \
  --env GITHUB_REPOSITORY=OWNER/REPOSITORY \
  --volume "$PWD:/workspace" \
  --workdir /workspace \
  github-repository-settings:local
```

`INPUT_TOKEN` must be set in the local environment. It needs repository administration write access; collaborator reconciliation also requires permission to manage repository collaborators.

## Set up as a GitHub Action

The workflow example at [`examples/workflows/apply-settings.yml`](examples/workflows/apply-settings.yml) previews manifest changes on pull requests and applies them after a push to `main`.

1. Add a settings manifest at `.github/repository-settings.yml`; start with [`examples/settings/basic.yml`](examples/settings/basic.yml) and adjust it for the repository.
2. Use `crflynn/github-repository-settings@v0` in the workflow to run the published action. It pulls `ghcr.io/crflynn/github-repository-settings:v0`, a multi-architecture image published on stable releases. After the first release, set GHCR package visibility to public so workflows in other repositories can pull it. The PR preview requires a stable release containing this feature; until then, an older image may fail instead of producing a plan.
3. Create the repository secret `REPOSITORY_SETTINGS_TOKEN` with `Administration: write` and the required write permissions. The workflow uses this token for both PR previews and applies.
4. Copy the workflow example to `.github/workflows/apply-settings.yml`, update its `main` branch filter if needed, and commit both files on a trusted branch.
5. Open a pull request to see the dry-run plan; merging the manifest changes to `main` triggers the real apply.

The example uses `pull_request_target` so the token is available for fork pull requests. The PR preview runs in dry-run mode, which allows GitHub API reads and blocks all other HTTP methods before they reach GitHub. The workflow checks out the pull request tree under `.pr` only to read its manifest; it does not run pull request code. Preserve this separation and never execute pull request-provided actions or scripts with secrets.

The action reads the settings file from the checked-out workspace, and its `settings-file` input defaults to `.github/repository-settings.yml`. See [`examples/settings/full.yml`](examples/settings/full.yml) for all supported components and review its high-impact settings before use.

If the manifest uses `webhooks[].secret_env`, create the referenced repository secret and map it in both Action steps' `env` blocks; the example maps `REPOSITORY_SETTINGS_WEBHOOK_SECRET`.

### Dry-run behavior

Set the action input `dry-run: true` to allow GitHub API reads, block every other HTTP method before it reaches GitHub, and report each planned write as its HTTP method and API path. Request bodies are not logged. The action's `dry-run` input defaults to `false`.

A dry-run still needs read permission for the API endpoints used by the configured components. The report lists write requests the current implementation would attempt; it is not a field-level diff or a guarantee that a real write will succeed. Some idempotent API calls may be listed even when GitHub would make no effective state change.

## Manifest schema

The manifest has optional `repository`, `actions`, `collaborators`, `teams`, `environments`, `rulesets`, `pages`, `autolinks`, `deploy_keys`, `interaction_limits`, `custom_properties`, and `webhooks` sections. Unknown keys and multiple YAML documents are rejected before API calls. Omitted sections and resources remain unchanged; fields omitted within an existing component are preserved where the API supports partial updates.

For resources that support `state`, `state: absent` declares that the named resource should not exist; applying the manifest removes it. Removing a resource from YAML alone does not delete it: omitted resources remain unchanged. After applying the deletion, you may remove the `state: absent` entry; later runs leave the resource unchanged.

Settings manifests are kept in [`examples/settings/`](examples/settings/), separate from workflow examples in [`examples/workflows/`](examples/workflows/): [`basic.yml`](examples/settings/basic.yml) shows a small configuration, and [`full.yml`](examples/settings/full.yml) demonstrates all supported components. `full.yml` includes high-impact examples such as public visibility, security controls, access grants, and restrictive rulesets. Review every value, replace usernames, team slugs, IDs, custom property names, and the sample deploy key, and provide webhook secrets through environment variables before applying. Interaction limits require a public repository. The test suite decodes each settings manifest in `examples/settings/` against the current schema.

This compact example demonstrates repository and Actions settings plus an autolink:
```yaml
repository:
  description: Managed repository
  visibility: private
  default_branch: main
  has_issues: true
  dependabot_security_updates: true
  secret_scanning_push_protection: true
  topics:
    - managed
    - go
actions:
  permissions:
    enabled: true
    allowed_actions: selected
    sha_pinning_required: true
    selected_actions:
      github_owned_allowed: true
      verified_allowed: true
      patterns_allowed:
        - acme/*
  workflow:
    default_workflow_permissions: read
    can_approve_pull_request_reviews: false
  retention:
    days: 30
autolinks:
  - key_prefix: JIRA-
    url_template: https://issues.example/browse/<num>
    is_alphanumeric: false
```

### Components

- `repository`: supports metadata and behavior fields `description`, `homepage`, `visibility` (`public`, `private`, `internal`), `default_branch`, `is_template`, `has_pull_requests`, `pull_request_creation_policy` (`all` or `collaborators_only`), topics, issue/wiki/project/discussion/download flags, merge methods, auto-merge, branch update/deletion, signoff, and archive status. Security fields include Dependabot security updates, vulnerability alerts, private vulnerability reporting, code security, secret scanning, and secret-scanning push protection. Fields are optional; explicit false and empty topic lists are preserved.
- `actions`: permissions include `enabled`, `allowed_actions` (`all`, `local_only`, or `selected`), `selected_actions`, and `sha_pinning_required`; workflow defaults include `default_workflow_permissions` (`read` or `write`) and `can_approve_pull_request_reviews`; `fork_pr` controls fork workflow execution, write tokens, secrets/variables, and approval for private repositories; `retention.days` sets artifact and log retention (1–400).
- `collaborators`: each item requires `username`; `permission` is one of `pull`, `triage`, `push`, `maintain`, or `admin`. `state` defaults to `present`. Set `state: absent` explicitly to remove a collaborator; omission never removes one.
- `teams`: each item requires an organization team `slug` and repository permission (`pull`, `triage`, `push`, `maintain`, or `admin`). Team access uses the repository owner as the organization; this section is for organization-owned repositories.
- `autolinks`: each item requires a unique `key_prefix`; present entries require `url_template`. `is_alphanumeric` is optional and preserves omission. Existing matching entries are left alone when equal; conflicting values fail because GitHub provides no update endpoint. `state: absent` explicitly removes a matching autolink; omitted entries remain untouched.
- `deploy_keys`: each item identifies a public SSH key by its key bytes; present entries require `title` and `read_only`. `state: absent` removes only the matching key. Existing keys with different title or read-only state fail because GitHub provides no update endpoint; undeclared keys are preserved. Never include private-key material.
- `interaction_limits`: `limit` is `existing_users`, `contributors_only`, or `collaborators_only`; optional `expiry` is `one_day`, `three_days`, `one_week`, `one_month`, or `six_months`. `state: absent` explicitly removes the restriction. GitHub supports these restrictions on public repositories; inherited organization/owner restrictions can make an update or delete fail with HTTP 409.
- `custom_properties`: each item names a property already defined by the organization and supplies a string or list of strings. Only declared values are updated; omitted properties are preserved and this component does not create definitions or delete values. Requires repository read permission and repository-level `edit custom property values` permission to write.
- `webhooks`: `url` is the identity; present hooks require non-empty `events`. Optional `active`, `content_type` (`json` or `form`), and `insecure_ssl` configure hook behavior. `secret_env` references an environment variable; literal secrets are rejected. Existing event/active changes require `secret_env` because GitHub does not return webhook secrets. Config-only updates preserve an existing secret. `state: absent` explicitly deletes a hook; omitted hooks remain unchanged.
- `environments`: each item requires `name`; optional `wait_timer`, `prevent_self_review`, `can_admins_bypass`, `reviewers` (GitHub `User` or `Team` IDs), and `deployment_branch_policy` fields preserve existing values when omitted. A `reviewers: []` explicitly clears reviewers. Set `state: absent` explicitly to delete an environment.
- `rulesets`: supports `branch`, `tag`, and `push` targets; `active`, `disabled`, or `evaluate` enforcement; ref-name conditions; optional-ID `bypass_actors`; and typed rules for branch/tag protection and push restrictions. Branch/tag rules include workflows, code scanning, code coverage, and Copilot code review in addition to the rules below. Push rules restrict file paths/extensions, maximum path length, and maximum file size. Push rules apply only to `push` targets; other rules apply only to `branch` or `tag`. `state: absent` explicitly deletes a named repository ruleset; unlisted rulesets remain untouched.
- `pages`: `enabled` is required. When enabled, `build_type` is `legacy` or `workflow`; legacy requires `source.branch` and `source.path` (`/` or `/docs`). CNAME, public visibility, and HTTPS enforcement are optional. `enabled: false` explicitly disables Pages.

**Looking up IDs.** `actor_id` and similar `id` fields require numeric GitHub IDs, not usernames or team slugs. Use `gh api /users/LOGIN --jq .id` for a user and `gh api /orgs/ORG/teams/TEAM_SLUG --jq .id` for a team. These IDs are used for environment reviewers, required reviewers, dismissal restrictions, and `Team` ruleset bypass actors.

For an `Integration` bypass actor, use the GitHub App's App ID, not its installation ID. `OrganizationAdmin` does not require an ID; `DeployKey` uses no ID. `RepositoryRole` uses a GitHub-assigned role ID, but the [REST API documentation](https://docs.github.com/en/rest/repos/rules) does not list a complete mapping, so do not guess. To inspect the IDs returned for an existing repository ruleset, run `gh api /repos/OWNER/REPO/rulesets --jq '.[] | select(.name == "RULESET_NAME") | .bypass_actors[] | [.actor_type, (.actor_id // "null")] | @tsv'`.

This action currently accepts `Integration`, `OrganizationAdmin`, `RepositoryRole`, `Team`, and `DeployKey` for `rulesets[].bypass_actors`; it does not yet accept GitHub's `User` bypass actor type.

- Branch/tag rules require `conditions.ref_name.include` with at least one entry and an explicit `exclude` list. Push rulesets do not accept `ref_name`. Each ruleset must contain at least one rule.

| Rule type | Required parameters | Optional parameters |
|---|---|---|
| `creation`, `deletion`, `required_linear_history`, `required_signatures`, `non_fast_forward` | None | None |
| `update` | `update_allows_fetch_and_merge` | None |
| `pull_request` | `required_approving_review_count` | `allowed_merge_methods`, `dismiss_stale_reviews_on_push`, `require_code_owner_review`, `require_last_push_approval`, `required_review_thread_resolution`, `dismissal_restriction`, `required_reviewers` |
| `required_status_checks` | `required_status_checks` (list of `context` and optional `integration_id`) | `strict_required_status_checks_policy`, `do_not_enforce_on_create` |
| `required_deployments` | `required_deployment_environments` | None |
| `merge_queue` | `check_response_timeout_minutes`, `grouping_strategy` (`ALLGREEN` or `HEADGREEN`), `max_entries_to_build`, `max_entries_to_merge`, `merge_method` (`MERGE`, `REBASE`, or `SQUASH`), `min_entries_to_merge`, `min_entries_to_merge_wait_minutes` | None |
| `workflows` | `workflows` (each requires `path` and either `sha` or `ref`) | `do_not_enforce_on_create` |
| `code_scanning` | `code_scanning_tools` | None |
| `code_coverage` | `max_coverage_drop` or `minimum_coverage` (0–100) | The other coverage field |
| `copilot_code_review` | At least one of `review_on_push`, `review_draft_pull_requests` | The other field |
| `file_path_restriction` | `restricted_file_paths` | None |
| `file_extension_restriction` | `restricted_file_extensions` | None |
| `max_file_path_length` | Positive `max_file_path_length` | None |
| `max_file_size` | Positive `max_file_size` | None |
| Pattern rules | `name`, `pattern`, `operator` (`starts_with`, `ends_with`, `contains`, or `regex`) | `negate` |

Pattern rules are `commit_message_pattern`, `commit_author_email_pattern`, `committer_email_pattern`, `branch_name_pattern`, and `tag_name_pattern`.

Repository ruleset `rules` map to the corresponding typed go-github rule parameter structures. The action manages repository rulesets only; inherited organization or enterprise rulesets are not updated.

## Development

Requires mise and Just. Run `mise install` to install the pinned Go and Just versions; `just setup` repeats this when tool versions change. Use `just test`, `just vet`, and `just build` individually, or `just check` for formatting, tests, vet, and build checks. Run `just image` to build the Dockerfile locally. The image workflow verifies `linux/amd64` and `linux/arm64` builds on branches and pull requests, then publishes stable releases to GHCR with version, major, minor, and `latest` tags.
