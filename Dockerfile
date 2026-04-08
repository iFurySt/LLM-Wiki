FROM --platform=$BUILDPLATFORM golang:1.26 AS build

ARG LLM_WIKI_VERSION=0.1.0-dev
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
    -ldflags "-s -w -X github.com/ifuryst/llm-wiki/internal/version.Version=${LLM_WIKI_VERSION}" \
    -o /out/llm-wiki-server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/llm-wiki-server /usr/local/bin/llm-wiki-server
COPY --from=build /src/install /app/install
COPY --from=build /src/skills /app/skills

EXPOSE 8234

ENTRYPOINT ["/usr/local/bin/llm-wiki-server"]
