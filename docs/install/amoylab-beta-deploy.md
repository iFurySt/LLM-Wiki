# Amoylab Beta Deploy

This is the current beta host flow for `/opt/llm-wiki`.

What it does:

- pushes to `main` publish `ghcr.io/ifuryst/llm-wiki:beta`
- GitHub Actions can SSH to amoylab and refresh the running container

Required repo assets:

- `.github/workflows/beta.yml`
- `deploy/prod/docker-compose.yml`
- `deploy/prod/.env.example`

Server prep:

```sh
sudo mkdir -p /opt/llm-wiki
sudo chown "$USER":"$USER" /opt/llm-wiki
cp deploy/prod/.env.example /opt/llm-wiki/.env.docker
```

Then set the real values in `/opt/llm-wiki/.env.docker`, especially:

- `LLM_WIKI_INSTALL_BASE_URL`
- PostgreSQL credentials

Deploy:

```sh
cd /opt/llm-wiki
docker compose pull
docker compose up -d
```

If the workflow deploys automatically, it performs the same refresh after pushing the new beta image.
