# CLI Usage

LLM-Wiki CLI is a thin wrapper over the HTTP API. Use it by default when an agent is working in a terminal.

Prefer this working loop:

1. Discover the current `ns`, folders, and existing documents first.
2. Look up the target page by slug when possible.
3. Update an existing document if it already represents the knowledge.
4. Create a new document only when there is no good existing home.

## Common Commands

```sh
llm-wiki system info --base-url http://127.0.0.1:8234
llm-wiki folder list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
llm-wiki folder create --base-url http://127.0.0.1:8234 --token dev-bootstrap-token --key projects --display-name Projects --visibility private
llm-wiki document list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
llm-wiki document get-by-slug --base-url http://127.0.0.1:8234 --token dev-bootstrap-token 1 launch-plan
```

The CLI also installs a `lw` alias with the same arguments:

```sh
lw system info --base-url http://127.0.0.1:8234
lw folder list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
lw document list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
```

Preferred longer-lived flow:

```sh
lw auth login --base-url http://127.0.0.1:8234
lw auth whoami --base-url http://127.0.0.1:8234
```

`lw auth login --device-code` always prints the approval URL and code. On local machines it also tries to open the browser unless `--no-open` is passed. Base URL and profile state can live in `~/.llm-wiki/config.json`. Only `auth login` accepts `--ns`; other commands use the stored token context.

## Create A Document

```sh
llm-wiki document create \
  --base-url http://127.0.0.1:8234 \
  --token dev-bootstrap-token \
  --folder-id 1 \
  --slug launch-plan \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nInitial draft." \
  --author-type agent \
  --author-id codex \
  --change-summary "create initial draft"
```

Use create when the knowledge does not already have a natural existing page.

Source-aware create paths are also available:

```sh
llm-wiki document create text \
  --base-url http://127.0.0.1:8234 \
  --token dev-bootstrap-token \
  --folder-id 1 \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nInitial draft."

llm-wiki document create file \
  --base-url http://127.0.0.1:8234 \
  --token dev-bootstrap-token \
  --folder-id 1 \
  --path ./notes/launch-plan.md

llm-wiki document create url \
  --base-url http://127.0.0.1:8234 \
  --token dev-bootstrap-token \
  --folder-id 1 \
  --url https://example.com/launch-plan
```

`document create file` and `document create url` keep the imported body in the document and store provenance in structured `source` metadata on the document and each revision.

For URL-specific routing, use:

```sh
llm-wiki document create x --base-url http://127.0.0.1:8234 --token dev-bootstrap-token --folder-id 1 --url https://x.com/openai/status/1
llm-wiki document create zhihu --base-url http://127.0.0.1:8234 --token dev-bootstrap-token --folder-id 1 --url https://www.zhihu.com/question/1
```

`document create xiaohongshu` is currently reserved as a manual-only path and will tell you to fall back to pasted text or a local file.

## Update A Document

```sh
llm-wiki document update 1 \
  --base-url http://127.0.0.1:8234 \
  --token dev-bootstrap-token \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nUpdated by LLM-Wiki skill." \
  --author-type agent \
  --author-id claude-code \
  --change-summary "refine plan"
```

Use update for the common case where the page already exists and the task adds new durable information.

## Suggested Session Loop

For normal agent work:

```sh
llm-wiki folder list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token
llm-wiki document list --base-url http://127.0.0.1:8234 --token dev-bootstrap-token --folder-id 1
llm-wiki document get-by-slug --base-url http://127.0.0.1:8234 --token dev-bootstrap-token 1 launch-plan
llm-wiki document update 1 \
  --base-url http://127.0.0.1:8234 \
  --token dev-bootstrap-token \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nRefined with latest implementation status." \
  --author-type agent \
  --author-id codex \
  --change-summary "capture latest implementation status"
```
