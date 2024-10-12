FROM golang:1.22-bullseye

WORKDIR /app

COPY ./caddy .

RUN go mod download

WORKDIR /tmp

COPY Caddyfile.sample ./Caddyfile

WORKDIR /app

RUN go build -o newcaddy ./cmd/caddy

ENTRYPOINT [ "./newcaddy", "run", "--config", "/tmp/Caddyfile" ]
