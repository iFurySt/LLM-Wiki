# CLI Usage

DocMesh CLI is a thin wrapper over the HTTP API. Use it by default when an agent is working in a terminal.

Prefer this working loop:

1. Discover the tenant, namespace, and existing documents first.
2. Look up the target page by slug when possible.
3. Update an existing document if it already represents the knowledge.
4. Create a new document only when there is no good existing home.

## Common Commands

```sh
docmesh system info --base-url http://127.0.0.1:8234
docmesh space list --base-url http://127.0.0.1:8234 --tenant default
docmesh namespace list --base-url http://127.0.0.1:8234 --tenant default
docmesh namespace create --base-url http://127.0.0.1:8234 --tenant default --key projects --display-name Projects --visibility tenant
docmesh document list --base-url http://127.0.0.1:8234 --tenant default
docmesh document get-by-slug --base-url http://127.0.0.1:8234 --tenant default 1 launch-plan
```

The CLI also installs a `dm` alias with the same arguments:

```sh
dm system info --base-url http://127.0.0.1:8234
dm namespace list --base-url http://127.0.0.1:8234 --tenant default
dm document list --base-url http://127.0.0.1:8234 --tenant default
```

## Create A Document

```sh
docmesh document create \
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
docmesh document update 1 \
  --base-url http://127.0.0.1:8234 \
  --tenant default \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nUpdated by DocMesh skill." \
  --author-type agent \
  --author-id claude-code \
  --change-summary "refine plan"
```

Use update for the common case where the page already exists and the task adds new durable information.

## Suggested Session Loop

For normal agent work:

```sh
docmesh namespace list --base-url http://127.0.0.1:8234 --tenant default
docmesh document list --base-url http://127.0.0.1:8234 --tenant default --namespace-id 1
docmesh document get-by-slug --base-url http://127.0.0.1:8234 --tenant default 1 launch-plan
docmesh document update 1 \
  --base-url http://127.0.0.1:8234 \
  --tenant default \
  --title "Launch Plan" \
  --content "# Launch Plan\n\nRefined with latest implementation status." \
  --author-type agent \
  --author-id codex \
  --change-summary "capture latest implementation status"
```
