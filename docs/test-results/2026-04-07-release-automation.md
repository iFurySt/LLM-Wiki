# 2026-04-07 Release Automation Validation

## Scope

Validated the first GitHub Releases automation and installer changes:

- tag-driven GitHub Actions release workflow
- CLI archive naming for GitHub Releases downloads
- installer script switched to GitHub Releases as the binary source

## Commands Run

```bash
./scripts/release/package-install.sh
sh -n install/install-cli.sh
go test ./...
```

## Results

- release packaging passed and produced release-ready archive names without embedding the version in the filename
- installer shell syntax check passed
- Go test suite passed after updating install-route expectations

## Produced Assets

- `dist/install/releases/llm-wiki_darwin_amd64.tar.gz`
- `dist/install/releases/llm-wiki_darwin_arm64.tar.gz`
- `dist/install/releases/llm-wiki_linux_amd64.tar.gz`
- `dist/install/releases/llm-wiki_linux_arm64.tar.gz`
- `dist/install/releases/llm-wiki_windows_amd64.zip`
- `dist/install/checksums.txt`
- `dist/install/version.txt`
