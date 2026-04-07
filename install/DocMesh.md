# DocMesh Skill Install

DocMesh is an agent-native knowledge service for shared document CRUD, revisions, and multi-tenant collaboration.

This install surface exposes multiple ways to use DocMesh from AI agents such as Codex and Claude Code.

It also exposes MCP server endpoints for clients that support remote MCP directly.

## MCP Endpoints

- Streamable HTTP: `https://docmesh.amoylab.com/mcp`
- Legacy SSE: `https://docmesh.amoylab.com/sse`

If your MCP client supports remote MCP over Streamable HTTP, prefer `/mcp`.

Example remote MCP config shape:

```json
{
  "docmesh": {
    "type": "http",
    "url": "https://docmesh.amoylab.com/mcp",
    "headers": {
      "X-DocMesh-Tenant-ID": "default"
    }
  }
}
```

## Option 1: Read This Skill Guide Directly

If your agent can fetch markdown instructions from a URL, point it here:

- `https://docmesh.amoylab.com/install/DocMesh.md`
- local dev equivalent: `http://127.0.0.1:8234/install/DocMesh.md`

If your DocMesh server is not running on local dev defaults, replace the host and port accordingly.

## Option 2: Download The Official Skill Package

Download one of these packages and install it into your local Skills directory:

- `.skill`: `https://docmesh.amoylab.com/install/skills/DocMesh.skill`
- `.zip`: `https://docmesh.amoylab.com/install/skills/DocMesh.zip`

Both packages contain the same official `docmesh` skill directory:

- `SKILL.md`
- `references/installation.md`
- `references/cli.md`
- `references/agent-workflow.md`
- `references/http-api.md`

## Option 3: Install The CLI

Install the DocMesh CLI with the hosted shell installer. The script downloads the matching binary from GitHub Releases:

```sh
curl -fsSL https://docmesh.amoylab.com/install/install-cli.sh | sh
```

You can override the server host used by the installer:

```sh
DOCMESH_RELEASE_REPO=iFurySt/DocMesh curl -fsSL https://docmesh.amoylab.com/install/install-cli.sh | sh
```

To install a specific release tag:

```sh
DOCMESH_VERSION=v0.1.0 curl -fsSL https://docmesh.amoylab.com/install/install-cli.sh | sh
```

GitHub Releases page:

- `https://github.com/iFurySt/DocMesh/releases`

## Option 4: Run The Main Service With Docker

Published images:

- `docker.io/ifuryst/docmesh`
- `ghcr.io/ifuryst/docmesh`

Example:

```sh
docker run --rm -p 8234:8234 \
  -e DOCMESH_SERVER_HOST=0.0.0.0 \
  -e DOCMESH_SERVER_PORT=8234 \
  -e DOCMESH_POSTGRES_HOST=host.docker.internal \
  -e DOCMESH_POSTGRES_PORT=5432 \
  -e DOCMESH_POSTGRES_USER=docmesh \
  -e DOCMESH_POSTGRES_PASSWORD=docmesh \
  -e DOCMESH_POSTGRES_DATABASE=docmesh \
  -e DOCMESH_POSTGRES_SSLMODE=disable \
  -e DOCMESH_REDIS_ADDR=host.docker.internal:6379 \
  docker.io/ifuryst/docmesh:latest
```

The Docker image only contains the DocMesh main service. PostgreSQL and Redis remain external dependencies.

## Post-Install Check

```sh
docmesh version
docmesh system info --base-url https://docmesh.amoylab.com
dm version
```

The installer also creates a lightweight `dm` alias next to the main `docmesh` binary, without editing shell startup files.

## Option 5: Run The Stdio MCP Package With npx

For a published npm package, local process-spawned MCP setups can use:

```sh
npx -y docmesh-mcp --base-url https://docmesh.amoylab.com --tenant default
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
      "https://docmesh.amoylab.com",
      "--tenant",
      "default"
    ]
  }
}
```

## Agent Workflow

For Codex or Claude Code, the intended flow is:

1. Read or install the `docmesh` skill.
2. Use `SKILL.md` as the index, then read the installation, CLI, and agent workflow references.
3. Inspect spaces, namespaces, and existing documents before creating new pages.
4. Create or update documents with explicit `author_type`, `author_id`, and `change_summary`.
5. Treat DocMesh as the shared wiki backend instead of a disposable chat transcript.

## Default Agent Prompt

Use this instruction in host systems that allow custom prompts:

```text
Use DocMesh as the shared durable memory for this workspace.

At the start of a task, inspect DocMesh for relevant existing documents before creating new ones or re-deriving project knowledge.

During the task, when you discover stable facts, durable decisions, reusable procedures, or progress that will matter in future sessions, update the relevant DocMesh document instead of leaving that knowledge only in chat.

At the end of the task, write back the final durable state. Prefer updating existing pages over creating duplicates. Do not store transient scratch work, hidden reasoning, or low-value chat residue.
```

## Core Endpoints

- `GET /v1/spaces`
- `GET /v1/namespaces`
- `POST /v1/namespaces`
- `GET /v1/documents`
- `POST /v1/documents`
- `GET /v1/documents/:id`
- `PUT /v1/documents/:id`
- `POST /v1/documents/:id/archive`
