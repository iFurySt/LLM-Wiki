FROM --platform=$BUILDPLATFORM golang:1.26 AS build

ARG DOCMESH_VERSION=0.1.0-dev
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -trimpath \
    -ldflags "-s -w -X github.com/ifuryst/docmesh/internal/version.Version=${DOCMESH_VERSION}" \
    -o /out/docmesh-server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/docmesh-server /usr/local/bin/docmesh-server
COPY --from=build /src/install /app/install
COPY --from=build /src/skills /app/skills

EXPOSE 8234

ENTRYPOINT ["/usr/local/bin/docmesh-server"]
