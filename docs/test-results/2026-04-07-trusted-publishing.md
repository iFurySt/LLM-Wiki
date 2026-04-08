# 2026-04-07 Trusted Publishing Validation

## Scope

Validated the final release distribution setup after repeated npm publish failures.

Focus areas:

- npm trusted publishing
- Docker release build duration and success
- end-to-end tag-driven release behavior

## Final Known-Good State

- npm publish workflow: `.github/workflows/publish-npm.yml`
- npm package: `@ifuryst/llm-wiki-mcp`
- npm trusted publisher:
  - repository: `iFurySt/LLM-Wiki`
  - workflow: `publish-npm.yml`
  - environment: empty
- npm publish runtime:
  - Node 24
  - npm 11.5.1+

## Observed Failure Modes During Bring-Up

- `E403`: npm token lacked publish-time 2FA bypass
- `E404`: OIDC provenance succeeded, but npm package authorization did not match the configured trusted publisher
- `ENEEDAUTH`: workflow runtime/auth setup did not satisfy trusted publishing requirements

## Fixes That Mattered

- moved npm publishing into a dedicated `publish-npm.yml`
- ensured the trusted publisher pointed to `publish-npm.yml`
- kept npm publishing on GitHub OIDC instead of `NPM_TOKEN`
- upgraded the publish workflow to Node 24 and npm 11.5.1+
- tracked `npm/llm-wiki-mcp/bin/llm-wiki-mcp.js` in git so publish artifacts were complete
- slimmed the production Dockerfile so release image builds no longer generated full CLI install bundles

## Final Result

- `publish-npm` workflow for tag `v0.1.8`: success
- published package version: `@ifuryst/llm-wiki-mcp@0.1.8`
- `npm dist-tag latest`: `0.1.8`
- Docker release workflow on the optimized path completed in low minutes instead of the original ~28 minute path

## Follow-Up

- delete the legacy `NPM_TOKEN` GitHub secret after confirming no remaining workflow uses it
- keep npm trusted publisher settings and workflow filename synchronized if the workflow is renamed again
