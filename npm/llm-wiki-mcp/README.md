# @ifuryst/llm-wiki-mcp

`@ifuryst/llm-wiki-mcp` is a stdio MCP server for LLM-Wiki. It is intended for local process-spawned integrations such as:

- Claude Code MCP config
- Codex-compatible MCP clients
- local agent runners that prefer `npx`

## Run

```bash
LLM_WIKI_TOKEN=<llm-wiki-token> npx -y @ifuryst/llm-wiki-mcp --base-url https://llm-wiki.ifuryst.com
```

Local dev equivalent:

```bash
LLM_WIKI_TOKEN=dev-bootstrap-token npx -y @ifuryst/llm-wiki-mcp --base-url http://127.0.0.1:8234
```

Environment variables:

- `LLM_WIKI_BASE_URL`
- `LLM_WIKI_TOKEN`

## Exposed Capabilities

- tools for listing, creating, updating, and archiving LLM-Wiki content
- resources for spaces, namespaces, and documents

The server is a thin stdio bridge over the LLM-Wiki HTTP API.
