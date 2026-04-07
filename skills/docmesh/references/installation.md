# Installation

Install the official DocMesh CLI from the running DocMesh server:

```sh
curl -fsSL http://127.0.0.1:8234/install/install-cli.sh | sh
```

If the server is not running on local defaults, override the base URL:

```sh
DOCMESH_BASE_URL=http://your-host:8234 \
  curl -fsSL http://your-host:8234/install/install-cli.sh | sh
```

After installation:

```sh
docmesh version
docmesh system info --base-url http://127.0.0.1:8234
```

The installer also places a short alias:

```sh
dm version
dm system info --base-url http://127.0.0.1:8234
```

## Skill Package

If your agent platform installs packaged skills, use either of these downloads:

- `http://127.0.0.1:8234/install/skills/DocMesh.skill`
- `http://127.0.0.1:8234/install/skills/DocMesh.zip`

Both archives contain the same `docmesh` skill directory.

## Hosted Guide

If the agent can read markdown instructions from a URL, point it to:

- `http://127.0.0.1:8234/install/DocMesh.md`

That hosted guide explains:

- how to install or connect through CLI, MCP, or `npx`
- how to use `SKILL.md` as the entry index
- how the agent should keep accumulating durable knowledge into DocMesh during normal tasks
