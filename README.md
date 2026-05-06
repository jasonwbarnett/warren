# warren

A small, funny web server running on a Raspberry Pi at [warren.barnett.network](https://warren.barnett.network).

It does nothing important. It does some entertaining things.

```
    .--.
   |o_o |    Server Bot v0.1
   |:_/ |    Status: vibing
  //   \ \   Hardware: Raspberry Pi
 (|     | )  Tunnel: cloudflared
/'\_   _/`\
\___)=(___/
```

## What's inside

| Path | Description |
|------|-------------|
| [`/`](https://warren.barnett.network/) | Home |
| [`/whoami`](https://warren.barnett.network/whoami) | What the server can see about you — IP, headers, and a deep psychological profile |
| [`/fortune`](https://warren.barnett.network/fortune) | Programming wisdom of dubious utility |
| [`/haiku`](https://warren.barnett.network/haiku) | The pain of software development, in 5-7-5 |
| [`/truth`](https://warren.barnett.network/truth) | Uncomfortable things no one wants to hear |
| [`/status`](https://warren.barnett.network/status) | Server metrics (some real, some aspirational) |
| [`/roast`](https://warren.barnett.network/roast) | Let the server judge your browser choices |
| [`/coffee`](https://warren.barnett.network/coffee) | RFC 2324 compliant — HTTP 418 I'm a Teapot |
| [`/ping`](https://warren.barnett.network/ping) | pong |
| [`/shrug`](https://warren.barnett.network/shrug) | ¯\_(ツ)_/¯ |
| [`/echo`](https://warren.barnett.network/echo) | POST something, get it back verbatim |
| [`/admin`](https://warren.barnett.network/admin) | Definitely a real admin panel |
| [`/sudo`](https://warren.barnett.network/sudo) | No |
| [`/robots.txt`](https://warren.barnett.network/robots.txt) | Has opinions about specific crawlers |

There are other things to find. The response headers are always interesting. Some paths reward persistence.

## Running it

```bash
go build -o server .
./server
```

Listens on `:8080` by default. Override with `PORT`:

```bash
PORT=3000 ./server
```

### As a systemd service

```ini
# /etc/systemd/system/warren.service
[Unit]
Description=warren web server
After=network.target

[Service]
ExecStart=/home/jason/server/server
WorkingDirectory=/home/jason/server
Restart=always
User=jason
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now warren
```

### With cloudflared

```bash
cloudflared tunnel run --url http://localhost:8080 warren
```

## Tech

- **Go** — stdlib only, zero dependencies
- **Raspberry Pi** — ARM64, Debian, 5W
- **Cloudflare Tunnel** — no open ports, no router config, no stress

## Easter eggs

There are a few things hidden in here. The source code is right in front of you, but where's the fun in that?
