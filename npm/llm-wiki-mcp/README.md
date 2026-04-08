# @ifuryst/llm-wiki-mcp

`@ifuryst/llm-wiki-mcp` is a stdio MCP server for LLM-Wiki. It is intended for local process-spawned integrations such as:

- Claude Code MCP config
- Codex-compatible MCP clients
- local agent runners that prefer `npx`

## Run

```bash
npx -y @ifuryst/llm-wiki-mcp --base-url https://llm-wiki.ifuryst.com --tenant default
```

Local dev equivalent:

```bash
npx -y @ifuryst/llm-wiki-mcp --base-url http://127.0.0.1:8234 --tenant default
```

Environment variables:

- `LLM_WIKI_BASE_URL`
- `LLM_WIKI_TENANT_ID`

## Exposed Capabilities

- tools for listing, creating, updating, and archiving LLM-Wiki content
- resources for spaces, namespaces, and documents

The server is a thin stdio bridge over the LLM-Wiki HTTP API.
