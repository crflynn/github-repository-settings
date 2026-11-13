# Agent instructions

- Keep Go toolchain versions in `mise.toml` and `Dockerfile` aligned; CI uses the version from `mise.toml`.
- Preserve the Go Docker Action contract in `action.yml` and `README.md` when changing inputs, manifest parsing, or API behavior.
- Run `go test ./...` and `go vet ./...` for Go changes. Build the action with `docker build -t github-repository-settings:local .` when changing the Docker image or executable.
- Keep CI split by concern in `.github/workflows/test.yml`, `lint.yml`, and `build-image.yml`; keep release automation in `release-please.yml` and validate Conventional Commit messages. Follow least-privilege permissions and pin third-party actions to commit SHAs.
- Maintain the release-please configuration and manifest with the `vX.Y.Z` release and moving `vX`/`vX.Y` Action tag conventions documented in `README.md`.
- Do not add secrets, repository-specific settings, or machine-specific paths to the project.

- Keep `examples/*.yml` current with the manifest schema and runtime behavior; update the examples whenever fields, components, validation, or API semantics change. Ensure examples remain generic and valid, and preserve `TestExampleManifestsDecode` coverage.
