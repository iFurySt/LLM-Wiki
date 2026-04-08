# CLI Usage

LLM-Wiki CLI is a thin wrapper over the HTTP API. Use it by default when an agent is working in a terminal.

Prefer this working loop:

1. Discover the tenant, namespace, and existing documents first.
2. Look up the target page by slug when possible.
3. Update an existing document if it already represents the knowledge.
4. Create a new document only when there is no good existing home.

## Common Commands

```sh
llm-wiki system info --base-url http://127.0.0.1:8234
llm-wiki space list --base-url http://127.0.0.1:8234 --tenant default
llm-wiki namespace list --base-url http://127.0.0.1:8234 --tenant default
llm-wiki namespace create --base-url http://127.0.0.1:8234 --tenant default --key projects --display-name Projects --visibility tenant
llm-wiki document list --base-url http://127.0.0.1:8234 --tenant default
llm-wiki document get-by-slug --base-url http://127.0.0.1:8234 --tenant default 1 launch-plan
```

The CLI also installs a `lw` alias with the same arguments:

```sh
lw system info --base-url http://127.0.0.1:8234
lw namespace list --base-url http://127.0.0.1:8234 --tenant default
lw document list --base-url http://127.0.0.1:8234 --tenant default
```

## Create A Document

```sh
llm-wiki document create \
  --base-url http://127.0.0.1:8234 \
  --tenant default \
  --namespace-id 1 \
  --slug launch-plan \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nInitial draft." \
  --author-type agent \
  --author-id codex \
  --change-summary "create initial draft"
```

Use create when the knowledge does not already have a natural existing page.

## Update A Document

```sh
llm-wiki document update 1 \
  --base-url http://127.0.0.1:8234 \
  --tenant default \
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
llm-wiki namespace list --base-url http://127.0.0.1:8234 --tenant default
llm-wiki document list --base-url http://127.0.0.1:8234 --tenant default --namespace-id 1
llm-wiki document get-by-slug --base-url http://127.0.0.1:8234 --tenant default 1 launch-plan
llm-wiki document update 1 \
  --base-url http://127.0.0.1:8234 \
  --tenant default \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nRefined with latest implementation status." \
  --author-type agent \
  --author-id codex \
  --change-summary "capture latest implementation status"
```
