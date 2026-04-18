# Deployment Notes

This repository combines three moving parts:

- the Go web application
- the MySQL/MariaDB schema
- the Asterisk/AMI side

## Minimal Bring-Up

1. Import the dialer schema:

   ```bash
   mysql -u root -p < install/dialer.sql
   ```

2. Create the runtime config:

   ```bash
   cp app/config_sample.json config.json
   ```

3. Update these values in `config.json`:

   - database credentials and host
   - `ami.host` and `oami.host`
   - `allowedips`
   - HTTP/HTTPS ports
   - TLS file paths if HTTPS is enabled

4. Build and run:

   ```bash
   go build -o bin/robocall ./app
   ./bin/robocall
   ```

## AMI Expectations

The app expects working AMI connectivity. In practice that means:

- an Asterisk host reachable from the app
- AMI users created with the permissions needed for queue and originate
- queue/channel activity available to the application

## Web UI Expectations

The UI renders templates from `templates/` and static assets from `static/`.
If you deploy the binary separately from the repository, keep those directories
next to the executable or copy them into the container image, as the included
Dockerfile does.

## Legacy Install Scripts

The shell scripts under `install/` are historical operator helpers for Asterisk
bootstrap and DB preparation. They are useful references, but they are not a
substitute for a controlled production build pipeline.

Review before use:

- third-party codec downloads
- package manager assumptions
- system file modifications
- service ownership and logrotate changes
