# docmesh-mcp

`docmesh-mcp` is a stdio MCP server for DocMesh. It is intended for local process-spawned integrations such as:

- Claude Code MCP config
- Codex-compatible MCP clients
- local agent runners that prefer `npx`

## Run

```bash
npx -y docmesh-mcp --base-url https://docmesh.amoylab.com --tenant default
```

Local dev equivalent:

```bash
npx -y docmesh-mcp --base-url http://127.0.0.1:8234 --tenant default
```

Environment variables:

- `DOCMESH_BASE_URL`
- `DOCMESH_TENANT_ID`

## Exposed Capabilities

- tools for listing, creating, updating, and archiving DocMesh content
- resources for spaces, namespaces, and documents

The server is a thin stdio bridge over the DocMesh HTTP API.
