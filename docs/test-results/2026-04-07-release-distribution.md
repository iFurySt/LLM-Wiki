# 2026-04-07 Release Distribution Validation

This record is the initial bootstrap checkpoint only.

For the final npm trusted publishing result and the corrected release path, see:

- `docs/test-results/2026-04-07-trusted-publishing.md`

## Scope

Validated the first repo-side implementation of tag-driven release distribution for:

- GitHub Releases
- Docker Hub
- GHCR
- npm

## Commands Run

```sh
go test ./...
npm pack --dry-run
LLM_WIKI_VERSION=v0.1.0 ./scripts/release/package-install.sh
docker build -f Dockerfile -t llm-wiki:test .
```

## Result

- `go test ./...`: passed
- `npm pack --dry-run`: passed for `@ifuryst/llm-wiki-mcp@0.1.0-dev`
- `package-install.sh`: passed and wrote `dist/install/version.txt` as `v0.1.0`
- `docker build`: passed for the production `Dockerfile`

## Notes

- The npm publish path is configured in workflow but was not executed locally.
- The actual registry pushes initially depended on GitHub Actions secrets.
- That npm token-based path was later replaced by npm trusted publishing over GitHub OIDC.
- Direct execution of the packaged Linux CLI binary was not validated on this macOS workstation because the host architecture and target binary format do not match.
