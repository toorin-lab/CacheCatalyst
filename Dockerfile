FROM golang:1.22-bullseye

WORKDIR /app

COPY ./caddy .

RUN go mod download

RUN go build -o newcaddy ./cmd/caddy

ENTRYPOINT [ "./newcaddy", "run", "--config", "/tmp/storage/Caddyfile" ]
