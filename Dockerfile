FROM golang:1.26 AS build

ARG DOCMESH_VERSION=0.1.0-dev

WORKDIR /src

RUN apt-get update && apt-get install -y zip unzip && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN chmod +x /src/scripts/release/package-install.sh /src/install/install-cli.sh
RUN DOCMESH_VERSION="${DOCMESH_VERSION}" /src/scripts/release/package-install.sh
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags "-X github.com/ifuryst/docmesh/internal/version.Version=${DOCMESH_VERSION}" \
  -o /out/docmesh-server ./cmd/server

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /out/docmesh-server /usr/local/bin/docmesh-server
COPY --from=build /src/install /app/install
COPY --from=build /src/skills /app/skills
COPY --from=build /src/dist/install /app/dist/install

EXPOSE 8234

CMD ["docmesh-server"]
