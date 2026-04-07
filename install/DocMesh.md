# DocMesh Skill Install

DocMesh is an agent-native knowledge service for shared document CRUD, revisions, and multi-tenant collaboration.

This install surface exposes multiple ways to use DocMesh from AI agents such as Codex and Claude Code.

It also exposes MCP server endpoints for clients that support remote MCP directly.

## MCP Endpoints

- Streamable HTTP: `http://127.0.0.1:8234/mcp`
- Legacy SSE: `http://127.0.0.1:8234/sse`

If your MCP client supports remote MCP over Streamable HTTP, prefer `/mcp`.

Example remote MCP config shape:

```json
{
  "docmesh": {
    "type": "http",
    "url": "http://127.0.0.1:8234/mcp",
    "headers": {
      "X-DocMesh-Tenant-ID": "default"
    }
  }
}
```

## Option 1: Read This Skill Guide Directly

If your agent can fetch markdown instructions from a URL, point it here:

- `http://127.0.0.1:8234/install/DocMesh.md`

If your DocMesh server is not running on local dev defaults, replace the host and port accordingly.

## Option 2: Download The Official Skill Package

Download one of these packages and install it into your local Skills directory:

- `.skill`: `http://127.0.0.1:8234/install/skills/DocMesh.skill`
- `.zip`: `http://127.0.0.1:8234/install/skills/DocMesh.zip`

Both packages contain the same official `docmesh` skill directory:

- `SKILL.md`
- `references/installation.md`
- `references/cli.md`
- `references/http-api.md`

## Option 3: Install The CLI

Install the DocMesh CLI with the hosted shell installer. The script downloads the matching binary from GitHub Releases:

```sh
curl -fsSL http://127.0.0.1:8234/install/install-cli.sh | sh
```

You can override the server host used by the installer:

```sh
DOCMESH_RELEASE_REPO=iFurySt/DocMesh curl -fsSL http://127.0.0.1:8234/install/install-cli.sh | sh
```

To install a specific release tag:

```sh
DOCMESH_VERSION=v0.1.0 curl -fsSL http://127.0.0.1:8234/install/install-cli.sh | sh
```

GitHub Releases page:

- `https://github.com/iFurySt/DocMesh/releases`

## Post-Install Check

```sh
docmesh version
docmesh system info --base-url http://127.0.0.1:8234
```

## Option 4: Run The Stdio MCP Package With npx

For a published npm package, local process-spawned MCP setups can use:

```sh
npx -y docmesh-mcp --base-url http://127.0.0.1:8234 --tenant default
```

Before the package is published, use the in-repo package directly:

```sh
npm install --prefix npm/docmesh-mcp --package-lock=false
npx --prefix npm/docmesh-mcp docmesh-mcp --base-url http://127.0.0.1:8234 --tenant default
```

Example stdio MCP config shape:

```json
{
  "docmesh": {
    "command": "npx",
    "args": [
      "-y",
      "docmesh-mcp",
      "--base-url",
      "http://127.0.0.1:8234",
      "--tenant",
      "default"
    ]
  }
}
```

## Agent Workflow

For Codex or Claude Code, the intended flow is:

1. Read or install the `docmesh` skill.
2. Use the CLI or HTTP API to inspect spaces, namespaces, and documents.
3. Create or update documents with explicit `author_type`, `author_id`, and `change_summary`.
4. Treat DocMesh as the shared wiki backend instead of a disposable chat transcript.

## Core Endpoints

- `GET /v1/spaces`
- `GET /v1/namespaces`
- `POST /v1/namespaces`
- `GET /v1/documents`
- `POST /v1/documents`
- `GET /v1/documents/:id`
- `PUT /v1/documents/:id`
- `POST /v1/documents/:id/archive`
