# HTTP API Usage

Use the HTTP API when the environment cannot run the CLI or when a tool wrapper wants direct JSON.

## Authentication

Every request should send a bearer token:

```text
Authorization: Bearer <llm-wiki-token>
```

## List Folders

```sh
curl -s http://127.0.0.1:8234/v1/namespaces \
  -H 'Authorization: Bearer <llm-wiki-token>'
```

## Create Folder

```sh
curl -s http://127.0.0.1:8234/v1/namespaces \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <llm-wiki-token>' \
  -d '{
    "key": "projects",
    "display_name": "Projects",
    "description": "shared project knowledge",
    "visibility": "private"
  }'
```

## Create Document

```sh
curl -s http://127.0.0.1:8234/v1/documents \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <llm-wiki-token>' \
  -d '{
    "namespace_id": 1,
    "slug": "launch-plan",
    "title": "Launch Plan",
    "content": "# Launch Plan\n\nInitial draft.",
    "author_type": "agent",
    "author_id": "codex",
    "change_summary": "create initial draft"
  }'
```

## Update Document

```sh
curl -s -X PUT http://127.0.0.1:8234/v1/documents/1 \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <llm-wiki-token>' \
  -d '{
    "title": "Launch Plan",
    "content": "# Launch Plan\n\nUpdated by agent.",
    "author_type": "agent",
    "author_id": "claude-code",
    "change_summary": "refine draft"
  }'
```
