# UI Agent Guide

This document defines the browser-side collaboration rules for `LLM-Wiki` UI work.

## Scope

- any change under the hosted web UI
- browser login and auth flows
- `/setup`, `/ui`, `/install/*`, `/admin/*`, and related user-visible behavior
- manual verification, debugging, and UI acceptance

## Browser Verification

- For UI-facing changes, browser verification is the default expectation unless the user explicitly says it is unnecessary.
- Use `skills/chrome-devtools-cli/` for reproduction, inspection, debugging, and acceptance.
- Use direct browser evidence such as DOM structure, computed styles, screenshots, console output, and network requests instead of code-only guesses.

## Browser Binary And Profile

- Always use Chrome stable for this repo.
- Do not switch to Canary, Beta, Dev, Chromium, or Chrome for Testing.
- Always reuse the repo profile at `.llmwiki/chrome-profile`.
- Do not use a random fresh profile or the system default profile for repo verification.
- Prefer `scripts/browser/open-chrome-stable.sh` and `scripts/browser/chrome-devtools-mcp.sh` or the matching `make` targets so these rules are enforced consistently.

## Profile Reuse Rules

- If `.llmwiki/chrome-profile` already exists, continue reusing it for browser debugging and acceptance.
- If `.llmwiki/chrome-profile` does not exist, create it first and use it from the first browser run onward.
- If another Chrome stable instance already owns that profile, reuse the existing instance instead of creating a second profile.
- If the existing instance already exposes a remote debugging port such as `http://127.0.0.1:9222`, connect the CLI to that running browser instead of relaunching Chrome.
- If the current login state is bound to `localhost`, keep using that host form instead of switching to `127.0.0.1` and breaking cookies.

## Login State Rules

- If the page redirects to a login flow, treat that as missing login state in the current `.llmwiki/chrome-profile`.
- In that case, launch headed Chrome stable with `.llmwiki/chrome-profile`, complete login once, then continue reusing the same profile.
- Do not create a second profile just to get through login.

## Verification Flow

1. Reproduce the issue or open the target page in Chrome stable with `.llmwiki/chrome-profile`.
2. Inspect the page through `scripts/browser/chrome-devtools-mcp.sh` using DOM, style, console, and network evidence.
3. Make the code change.
4. Re-run the same browser checks to confirm the change and catch visible regressions.

## Default Targets

- setup: `http://127.0.0.1:8234/setup`
- wiki UI: `http://127.0.0.1:8234/ui`
- install UI: `http://127.0.0.1:8234/install/LLM-Wiki.md`
- admin login: `http://127.0.0.1:8234/admin/login`
