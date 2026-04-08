# MCP Usage

LLM-Wiki exposes MCP in two forms:

- remote MCP over Streamable HTTP at `http://127.0.0.1:8234/mcp`
- local stdio MCP via `npx -y @ifuryst/llm-wiki-mcp`

## Remote MCP

Prefer this when the agent runtime supports remote MCP servers.

Recommended endpoint:

```text
http://127.0.0.1:8234/mcp
```

Legacy compatibility endpoint:

```text
http://127.0.0.1:8234/sse
```

Send the tenant header when the client allows custom headers:

```text
X-LLM-Wiki-Tenant-ID: default
```

## npx stdio MCP

Prefer this when the runtime expects a local process-spawned MCP server.

```sh
npx -y @ifuryst/llm-wiki-mcp --base-url http://127.0.0.1:8234 --tenant default
```

The package is a thin stdio bridge over the LLM-Wiki HTTP API and exposes the same LLM-Wiki-oriented tools and resources.
