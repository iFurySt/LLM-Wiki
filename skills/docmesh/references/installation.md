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

## Skill Package

If your agent platform installs packaged skills, use either of these downloads:

- `http://127.0.0.1:8234/install/skills/DocMesh.skill`
- `http://127.0.0.1:8234/install/skills/DocMesh.zip`

Both archives contain the same `docmesh` skill directory.
