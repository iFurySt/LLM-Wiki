# 2026-04-07 Release Distribution Validation

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
DOCMESH_VERSION=v0.1.0 ./scripts/release/package-install.sh
docker build -f Dockerfile -t docmesh:test .
```

## Result

- `go test ./...`: passed
- `npm pack --dry-run`: passed for `docmesh-mcp@0.1.0-dev`
- `package-install.sh`: passed and wrote `dist/install/version.txt` as `v0.1.0`
- `docker build`: passed for the production `Dockerfile`

## Notes

- The npm publish path is configured in workflow but was not executed locally.
- The actual registry pushes still depend on GitHub Actions secrets:
  - `DOCKERHUB_USERNAME`
  - `DOCKERHUB_TOKEN`
  - `NPM_TOKEN`
- Direct execution of the packaged Linux CLI binary was not validated on this macOS workstation because the host architecture and target binary format do not match.
