FROM golang:1.26 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /docmesh-server ./cmd/server

FROM debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /docmesh-server /usr/local/bin/docmesh-server

EXPOSE 8234

CMD ["docmesh-server"]

