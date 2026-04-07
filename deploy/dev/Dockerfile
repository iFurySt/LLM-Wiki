FROM golang:1.26

WORKDIR /workspace

RUN apt-get update && apt-get install -y ca-certificates zip unzip && rm -rf /var/lib/apt/lists/*
RUN go install github.com/air-verse/air@v1.63.1

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8234

CMD ["air", "-c", ".air.toml"]
