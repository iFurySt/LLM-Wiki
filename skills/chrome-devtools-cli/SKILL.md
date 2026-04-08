---
name: chrome-devtools-cli
description: Use this skill to drive browser tasks from the terminal through Chrome DevTools CLI while reusing the repo browser profile.
---

The `chrome-devtools-cli` skill lets you inspect and automate browser behavior from the terminal.

## Setup

If this is the first time using the CLI, see [references/installation.md](references/installation.md). Installation is a one-time prerequisite, not part of normal task flow.

## Repo Rules

- Always use Chrome stable for this repo. Do not switch to Canary, Beta, Dev, Chromium, or Chrome for Testing.
- Reuse the repo browser profile at `.llmwiki/chrome-profile`.
- If the profile already exists and is in use by a running Chrome stable instance, connect to that instance instead of creating a new profile.
- If login state is missing, use the same `.llmwiki/chrome-profile` in a headed Chrome stable session, complete login once, then continue reusing that profile.
- Prefer the repo wrappers `scripts/browser/open-chrome-stable.sh` and `scripts/browser/chrome-devtools-mcp.sh` over ad hoc raw CLI invocation.

## AI Workflow

1. Start from `scripts/browser/chrome-devtools-mcp.sh` or `make browser-mcp` so the repo profile and Chrome stable constraints are applied.
2. Use `take_snapshot` to identify elements.
3. Use `click`, `fill`, `press_key`, screenshots, console, and network tools as evidence.
4. Re-run the same checks after the fix or change.

## Command Usage

```sh
./scripts/browser/chrome-devtools-mcp.sh [flags]
```

For headed login or manual recovery, use:

```sh
./scripts/browser/open-chrome-stable.sh
```

Use `chrome-devtools --help` or `chrome-devtools-mcp --help` for low-level command details when needed.

## Common Commands

```bash
chrome-devtools list_pages
chrome-devtools new_page "https://example.com"
chrome-devtools navigate_page --url "https://example.com"
chrome-devtools take_snapshot
chrome-devtools click "uid"
chrome-devtools fill "uid" "text"
chrome-devtools press_key "Enter"
chrome-devtools list_console_messages
chrome-devtools list_network_requests
chrome-devtools take_screenshot
chrome-devtools evaluate_script "() => document.title"
```

## Service Management

```bash
chrome-devtools start
chrome-devtools status
chrome-devtools stop
```
