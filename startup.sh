#!/usr/bin/env bash
sudo mkdir /app
sudo chmod /app 777
mkdir /app/files

python3 shell.py -media-dir /app/files -prefix /
cd caddy || exit
go mod tidy
go build -o caddy ./cmd/caddy
chmod +x ./caddy
cp ../Caddyfile.sample Caddyfile
sudo ./caddy run --config Caddyfile