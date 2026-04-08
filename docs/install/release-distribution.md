# Release Distribution

Two release paths exist:

- `main` pushes
- version tags like `v0.1.0`

`main` publishes:

- `ghcr.io/ifuryst/llm-wiki:beta`
- optional beta deploy to amoylab

Version tags publish:

- GitHub Release assets
- Docker images to Docker Hub and GHCR
- npm package `@ifuryst/llm-wiki-mcp`

GitHub Release assets include:

- CLI archives
- skill archives
- checksums and version files
- `install/install-cli.sh`
- `install/LLM-Wiki.md`

Notes:

- the Docker image only contains the main LLM-Wiki service
- PostgreSQL is still the operator-managed dependency
- the installer downloads CLI binaries from GitHub Releases
