FROM golang:1.26 AS build

WORKDIR /app

RUN apt-get update && apt-get install -y zip unzip && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN chmod +x /app/scripts/release/package-install.sh /app/install/install-cli.sh
RUN /app/scripts/release/package-install.sh
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /docmesh-server ./cmd/server

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /docmesh-server /usr/local/bin/docmesh-server
COPY --from=build /app/install /app/install
COPY --from=build /app/dist/install /app/dist/install

EXPOSE 8234

CMD ["docmesh-server"]
