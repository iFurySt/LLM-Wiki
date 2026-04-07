# Release Distribution

This document records what happens on a pushed tag such as `v0.1.0`.

## Tag Trigger

The release workflow triggers on:

```text
refs/tags/v*
```

## Published Outputs

### GitHub Releases

Published by `softprops/action-gh-release`:

- CLI archives from `dist/install/releases/*`
- skill archives from `dist/install/skills/*`
- `dist/install/checksums.txt`
- `dist/install/version.txt`
- `install/install-cli.sh`
- `install/DocMesh.md`

### Docker Registries

Published multi-arch service image:

- `docker.io/ifuryst/docmesh`
- `ghcr.io/ifuryst/docmesh`

Tag shapes:

- exact version, for example `0.1.0`
- minor line, for example `0.1`
- major line, for example `0`
- `latest`

### npm

Published package:

- `docmesh-mcp`

The workflow rewrites `npm/docmesh-mcp/package.json` from `0.1.0-dev` to the pushed git tag version before publish.

## Required GitHub Secrets

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`
- `NPM_TOKEN`

`NPM_TOKEN` must be able to publish packages with 2FA bypass enabled. A plain read/write token without publish bypass will fail with npm `E403`.

## Permission Model

- GitHub Releases uses `contents: write`
- GHCR publish uses `packages: write` with `GITHUB_TOKEN`
- Docker Hub publish uses `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`
- npm publish uses `NPM_TOKEN`
- if npm account policy enforces publish-time 2FA, the token must explicitly support bypass, or the repo must move to npm trusted publishing

## Notes

- The Docker image only packages the DocMesh main service.
- PostgreSQL and Redis remain external dependencies supplied by the operator.
- The installer script downloads CLI archives from GitHub Releases instead of from the running DocMesh service.
