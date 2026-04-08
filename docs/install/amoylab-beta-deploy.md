# Amoylab Beta Deploy

This document records the intended deployment path for the amoylab LLM-Wiki host at `/opt/llm-wiki`.

The target state is:

- pushes to `main` publish `ghcr.io/ifuryst/llm-wiki:beta`
- GitHub Actions SSHes into amoylab and refreshes the running service from that image
- the host runs LLM-Wiki through Docker Compose instead of a hand-managed `systemd` process
- old unused LLM-Wiki images are pruned after each deploy so the server does not accumulate layers indefinitely

## Repo-Side Assets

The repository now provides:

- `.github/workflows/beta.yml`
- `deploy/prod/docker-compose.yml`
- `deploy/prod/.env.example`

`deploy/prod/docker-compose.yml` expects a real `/opt/llm-wiki/.env.docker` on the server and runs:

- image: `ghcr.io/ifuryst/llm-wiki:beta`
- container name: `llm-wiki`
- host network mode so the container reuses the server's loopback and existing local dependencies

That setup assumes a reverse proxy or tunnel is already terminating TLS for `https://llm-wiki.ifuryst.com/`.

## GitHub Secrets Needed

The automatic deploy job only runs when these repository secrets exist:

- `AMOYLAB_HOST`
- `AMOYLAB_USER`
- `AMOYLAB_SSH_KEY`

Optional, only if GHCR needs authenticated pulls from the server:

- `AMOYLAB_GHCR_USERNAME`
- `AMOYLAB_GHCR_TOKEN`

If the package stays public on GHCR, the optional pull credentials can remain unset.

## Server Preparation

Install or confirm:

- Docker Engine with the Compose v2 plugin
- a writable `/opt/llm-wiki`
- a populated `/opt/llm-wiki/.env.docker`

Recommended first-time bootstrap:

```sh
sudo mkdir -p /opt/llm-wiki
sudo chown "$USER":"$USER" /opt/llm-wiki
cp deploy/prod/.env.example /opt/llm-wiki/.env.docker
```

Then edit `/opt/llm-wiki/.env.docker` with the real production values, especially:

- `LLM_WIKI_INSTALL_BASE_URL`
- PostgreSQL credentials
- Redis address
- MinIO credentials
- OpenSearch URL

## Cutover From systemd

If the host is still running a native `llm-wiki` process under `systemd`, cut over once:

```sh
sudo systemctl stop llm-wiki
sudo systemctl disable llm-wiki
cd /opt/llm-wiki
docker compose pull
docker compose up -d
```

At that point the container restart policy becomes the service supervisor.

If you want to keep the old unit around for rollback, stop and disable it but do not delete it yet.

## Deploy Behavior

On every push to `main`, the beta workflow:

1. runs `go test ./...`
2. runs `npm pack --dry-run` for `npm/llm-wiki-mcp`
3. builds and pushes `ghcr.io/ifuryst/llm-wiki:beta`
4. uploads the production compose files to `/opt/llm-wiki`
5. runs `docker compose pull && docker compose up -d`
6. runs `docker image prune -af --filter "label=org.opencontainers.image.source=https://github.com/iFurySt/LLM-Wiki"`

The prune step removes unused old LLM-Wiki images after the new container is up.

## Manual Recovery

If the automatic SSH deploy fails, the same host can be refreshed manually with:

```sh
cd /opt/llm-wiki
docker compose pull
docker compose up -d
docker image prune -af --filter "label=org.opencontainers.image.source=https://github.com/iFurySt/LLM-Wiki"
```
