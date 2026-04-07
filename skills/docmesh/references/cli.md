# CLI Usage

DocMesh CLI is a thin wrapper over the HTTP API. Use it by default when an agent is working in a terminal.

## Common Commands

```sh
docmesh system info --base-url http://127.0.0.1:8234
docmesh space list --base-url http://127.0.0.1:8234 --tenant default
docmesh namespace list --base-url http://127.0.0.1:8234 --tenant default
docmesh namespace create --base-url http://127.0.0.1:8234 --tenant default --key projects --display-name Projects --visibility tenant
docmesh document list --base-url http://127.0.0.1:8234 --tenant default
docmesh document get-by-slug --base-url http://127.0.0.1:8234 --tenant default 1 launch-plan
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
