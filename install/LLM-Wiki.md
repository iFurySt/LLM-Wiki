# LLM-Wiki

LLM-Wiki is a shared knowledge service for agents.

Use it when you want one durable backend for documents, revisions, and shared agent-readable knowledge.

## Connect

Remote MCP:

- `https://llm-wiki.ifuryst.com/mcp`
- legacy SSE: `https://llm-wiki.ifuryst.com/sse`

Example:

```json
{
  "llm-wiki": {
    "type": "http",
    "url": "https://llm-wiki.ifuryst.com/mcp",
    "headers": {
      "Authorization": "Bearer <llm-wiki-token>"
    }
  }
}
```

## Install CLI

```sh
curl -fsSL https://llm-wiki.ifuryst.com/install/install-cli.sh | sh
```

Check:

```sh
lw version
lw auth login --base-url https://llm-wiki.ifuryst.com
lw folder list --base-url https://llm-wiki.ifuryst.com
```

## Run npx MCP

```sh
npx -y @ifuryst/llm-wiki-mcp --base-url https://llm-wiki.ifuryst.com
```

## Docker

```sh
docker run --rm -p 8234:8234 \
  -e LLM_WIKI_SERVER_HOST=0.0.0.0 \
  -e LLM_WIKI_SERVER_PORT=8234 \
  -e LLM_WIKI_POSTGRES_HOST=host.docker.internal \
  -e LLM_WIKI_POSTGRES_PORT=5432 \
  -e LLM_WIKI_POSTGRES_USER=llmwiki \
  -e LLM_WIKI_POSTGRES_PASSWORD=llmwiki \
  -e LLM_WIKI_POSTGRES_DATABASE=llmwiki \
  -e LLM_WIKI_POSTGRES_SSLMODE=disable \
  docker.io/ifuryst/llm-wiki:latest
```

## Agent Workflow

1. Inspect the current `ns`, folders, and documents first.
2. Reuse existing pages before creating new ones.
3. Write durable knowledge back into LLM-Wiki.
4. Avoid storing low-value scratch output as durable knowledge.

## Main Endpoints

- `GET /v1/namespaces`
- `POST /v1/namespaces`
- `GET /v1/documents`
- `POST /v1/documents`
- `GET /v1/documents/:id`
- `PUT /v1/documents/:id`
- `POST /v1/documents/:id/archive`
