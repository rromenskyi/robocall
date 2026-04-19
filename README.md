# robocall

Outbound autodialer web application for Asterisk, written in Go.

It manages call tasks in MySQL, serves a small web UI with Gin, connects to
Asterisk over AMI/OAMI, and feeds outbound calls into queues with configurable
limits.

## What This Project Is

`robocall` is a practical internal dialer-style service for bulk outbound
calling. It is not a generic telephony framework and not a hosted SaaS product.
It is a self-managed application that sits next to your Asterisk stack and
helps operators:

- upload call batches
- organize branches and geographies
- manage user access
- push calls into the dial queue
- observe queue state from Asterisk

## Main Components

- Go web app in [`app/`](app)
- HTML templates in [`templates/`](templates)
- Static assets in [`static/`](static)
- SQL and install helpers in [`install/`](install)
- Test helpers in [`test/`](test)

## How It Works

1. The web UI accepts tasks and CSV uploads
2. Tasks are expanded into queueable call rows in MySQL
3. Background workers process tasks and call queues
4. The app connects to Asterisk over AMI/OAMI
5. Queue and channel events are used to track call progress and limits

## Requirements

- Go 1.22+
- MySQL or MariaDB
- Asterisk with AMI enabled
- `config.json` in the repository root

Optional:

- TLS certificate/key files if you want HTTPS directly from the app
- Docker if you want to containerize the web service

## Quick Start

1. Prepare the database schema.

   ```bash
   mysql -u root -p < install/dialer.sql
   ```

2. Create a local config file.

   ```bash
   cp app/config_sample.json config.json
   ```

3. Adjust database credentials, AMI hosts, ports, allowed IPs, and TLS paths in
   `config.json`.

4. Build and run the app.

   ```bash
   go build -o bin/robocall ./app
   ./bin/robocall
   ```

By default the sample config serves HTTP on `:8080`.

## Configuration

The app loads configuration from `config.json` in the repository root.

Important fields:

- `global.httpport`: HTTP listen address, for example `:8080`
- `global.httpsport`: HTTPS listen address, for example `:443`
- `global.ssl_privatekey`: private key path
- `global.ssl_fullchain`: certificate chain path
- `global.allowedips`: comma-separated IP allowlist for the web UI
- `database.*`: MySQL/MariaDB connection settings
- `ami.*`: Asterisk AMI connection
- `oami.*`: second AMI/OAMI connection used for originate and queue work

The session cookie secret can be overridden with:

```bash
export ROBOCALL_SESSION_SECRET="change-me"
```

## Docker

Build the container:

```bash
docker build -t robocall .
```

Run it with your local config mounted into `/app/config.json`:

```bash
docker run --rm -p 8080:8080 \
  -e ROBOCALL_SESSION_SECRET=change-me \
  -v "$(pwd)/config.json:/app/config.json:ro" \
  robocall
```

The container now exits early with a clear error if `/app/config.json` is not
mounted. `app/config_sample.json` is bundled only as a reference file inside the
image; the application still expects a real runtime config at `/app/config.json`.

For a local all-in-one stack with MariaDB and preloaded schema:

```bash
docker compose up --build
```

That stack uses [`compose.yml`](compose.yml), a runtime config at [`compose/config.json`](compose/config.json), and an init schema at [`compose/initdb/01-schema.sql`](compose/initdb/01-schema.sql). It brings up:

- `db`: MariaDB with the dialer and CDR tables
- `robocall`: the web app on `http://127.0.0.1:8080`

Local compose defaults:

- database: `asteriskcdrdb`
- database user: `dialer_user`
- database password: `dialer_pass`
- web login: `admin`
- web password: `admin`

`compose/config.json` intentionally sets `allowedips` to `*` for local development. Keep a real allowlist in production configs.

The app also exposes `GET /ping`, which is used for the container health check in both `Dockerfile` and `compose.yml`.

For a fuller local dev stack with MariaDB and a minimal Asterisk/AMI side:

```bash
docker compose -f compose.asterisk.yml up --build
```

That variant adds:

- `asterisk`: a local Asterisk container with AMI enabled on `5038`
- `robocall`: wired to `asterisk:5038` for both `ami` and `oami`
- a seeded IVR name `demo-ivr` and seeded geo `Local Demo`

The full dev stack serves the UI on `http://127.0.0.1:8081` to avoid colliding with the simpler `compose.yml` stack on `:8080`.

The bundled Asterisk profile is intentionally minimal: it is good for local AMI/OAMI bring-up and dialer flow testing, but it is not a production telephony layout.

## Published Images

GitHub Actions now publishes container images to:

- `ghcr.io/rromenskyi/robocall`
- `${DOCKER_USERNAME}/robocall` on Docker Hub when `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets are configured

Published tags include:

- `latest` on the default branch
- branch tags such as `main`
- git tags such as `v1.2.3`
- immutable commit tags such as `sha-<commit>`

## Asterisk Setup Notes

The repository contains legacy helper scripts under [`install/`](install) for
building Asterisk Certified and preparing the dialer database on CentOS/Amazon
Linux style systems.

Those scripts are useful as operator references, but they should be reviewed
before production use:

- [`install/asterisk-build.sh`](install/asterisk-build.sh)
- [`install/asterisk-postinstall.sh`](install/asterisk-postinstall.sh)
- [`install/dialer.sql`](install/dialer.sql)
- [`install/cdr.sql`](install/cdr.sql)

There is also a longer deployment note here:

- [docs/deployment.md](docs/deployment.md)

## Security Notes

- `config.json` contains credentials and should not be committed
- the web UI is gated by session auth and an IP allowlist
- set a real `ROBOCALL_SESSION_SECRET` outside development
- review public routes before exposing the service outside a trusted network

## Repository Layout

```text
.
├── app/                  Go application code
├── install/              SQL and Asterisk bootstrap helpers
├── static/               CSS and static files
├── templates/            Gin HTML templates
├── test/                 Test helpers and sample configs
├── Dockerfile            Container build
└── README.md
```
